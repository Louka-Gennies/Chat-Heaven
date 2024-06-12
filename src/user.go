package chatHeaven

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)

func UserHandler(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	if username == "" {
		http.Error(w, "User not specified", http.StatusBadRequest)
		return
	}

	if r.Method == "GET" {
		var email, profilePicture, createdAt, first_name, last_name string
		query := `SELECT email, profile_picture, first_name, last_name, createdAt FROM users WHERE username = ?`
		err := db.QueryRowContext(context.Background(), query, username).Scan(&email, &profilePicture, &first_name, &last_name, &createdAt)
		if err != nil {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}

		data := struct {
			Username       string
			Email          string
			ProfilePicture string
			FirstName      string
			LastName       string
			Posts          []Post
			Topics         []Topic
			CreatedAt      string
		}{
			Username:       username,
			Email:          email,
			ProfilePicture: profilePicture,
			FirstName:      first_name,
			LastName:       last_name,
			Posts:          getPostsByUser(username),
			Topics:         getTopicByUser(username),
			CreatedAt:      createdAt,
		}

		tmpl, err := template.ParseFiles("templates/user.html")
		if err != nil {
			http.Error(w, "Error reading the HTML file", http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, data)
	} else if r.Method == "POST" {
		file, handler, err := r.FormFile("profile_picture")
		if err != nil {
			http.Error(w, "Error during file upload", http.StatusInternalServerError)
			return
		}
		defer file.Close()

		os.MkdirAll("static/uploads", os.ModePerm)

		filePath := filepath.Join("static/uploads", handler.Filename)
		f, err := os.Create(filePath)
		if err != nil {
			http.Error(w, "Error saving the file", http.StatusInternalServerError)
			return
		}
		defer f.Close()
		io.Copy(f, file)

		updateSQL := `UPDATE users SET profile_picture = ? WHERE username = ?`
		_, err = db.ExecContext(context.Background(), updateSQL, "/static/uploads/"+handler.Filename, username)
		if err != nil {
			http.Error(w, "Error updating the profile picture", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/user?username=%s", username), http.StatusSeeOther)
	}
}

func UpdateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		username := r.URL.Query().Get("username")
		firstName := r.FormValue("first_name")
		lastName := r.FormValue("last_name")

		updateSQL := `UPDATE users SET first_name = ?, last_name = ? WHERE username = ?`
		_, err := db.ExecContext(context.Background(), updateSQL, firstName, lastName, username)
		if err != nil {
			http.Error(w, "Error updating the user", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/user?username=%s", username), http.StatusSeeOther)
	}
}

func addUser(username, email, motDePasse, profilePicture, date string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(motDePasse), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = db.ExecContext(context.Background(), `INSERT INTO users (username, email, mot_de_passe, profile_picture, first_name, last_name, createdAt) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		username, email, hashedPassword, profilePicture, "", "", date)
	if err != nil {
		return err
	}
	return nil
}

func verifyUser(username, motDePasse string) error {
	var motDePasseDB string
	err := db.QueryRowContext(context.Background(), "SELECT mot_de_passe FROM users WHERE username = ?", username).Scan(&motDePasseDB)
	if err != nil {
		return err
	}

	err = bcrypt.CompareHashAndPassword([]byte(motDePasseDB), []byte(motDePasse))
	if err != nil {
		return fmt.Errorf("incorrect password")
	}
	return nil
}