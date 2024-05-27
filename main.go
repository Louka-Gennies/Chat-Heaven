package main

import (
	"context"
	"database/sql"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)

var db *sql.DB
var store = sessions.NewCookieStore([]byte("menu-classique-burger"))

func main() {
	dbPath := "chatHeaven.db"

	var err error
	db, err = sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.ExecContext(context.Background(), `CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        username TEXT NOT NULL UNIQUE,
        email TEXT NOT NULL UNIQUE,
        mot_de_passe TEXT NOT NULL,
        profile_picture TEXT,
        user_likes INTEGER
    )`)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.ExecContext(context.Background(), `CREATE TABLE IF NOT EXISTS topics (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL UNIQUE,
		description TEXT NOT NULL,
		picture TEXT,
		user TEXT NOT NULL,
		topic_likes INTEGER,
		FOREIGN KEY (user) REFERENCES users(username)
	)`)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.ExecContext(context.Background(), `CREATE TABLE IF NOT EXISTS posts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL UNIQUE,
		content TEXT NOT NULL UNIQUE,
		picture TEXT,
		user TEXT NOT NULL,
		post_likes INTEGER,
		FOREIGN KEY (user) REFERENCES users(username)
	)`)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/user", userHandler)
	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/create-post", createPost)
	http.HandleFunc("/posts", postsHandler)
	http.HandleFunc("/topics", topicsHandler)
	http.HandleFunc("/create-topic", createTopic)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	log.Println("Server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	username, ok := session.Values["username"]

	if !ok {
		tmpl, err := template.ParseFiles("templates/home.html")
		if err != nil {
			http.Error(w, "Erreur de lecture du fichier HTML", http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, nil)
		return
	}

	var profilePicture string
	err := db.QueryRowContext(context.Background(), "SELECT profile_picture FROM users WHERE username = ?", username).Scan(&profilePicture)
	if err != nil {
		http.Error(w, "Erreur lors de la récupération de la photo de profil", http.StatusInternalServerError)
		return
	}

	data := struct {
		Username       string
		ProfilePicture string
	}{
		Username:       username.(string),
		ProfilePicture: profilePicture,
	}

	tmpl, err := template.ParseFiles("templates/home.html")
	if err != nil {
		http.Error(w, "Erreur de lecture du fichier HTML", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, data)
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		email := r.FormValue("email")
		password := r.FormValue("password")

		err := ajouterUtilisateur(username, email, password, "./static/uploads/blank-pfp.png")
		if err != nil {
			http.Error(w, "Erreur lors de l'inscription", http.StatusInternalServerError)
			return
		}

		session, _ := store.Get(r, "session")
		session.Values["username"] = username
		session.Save(r, w)

		http.Redirect(w, r, fmt.Sprintf("/user?username=%s", username), http.StatusSeeOther)

		return
	}

	tmpl, err := template.ParseFiles("templates/register.html")
	if err != nil {
		http.Error(w, "Erreur de lecture du fichier HTML", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		http.ServeFile(w, r, "templates/login.html")
	} else if r.Method == "POST" {
		username := r.FormValue("username")
		password := r.FormValue("password")

		err := verifierUtilisateur(username, password)
		if err != nil {
			http.Error(w, "Nom d'utilisateur ou mot de passe incorrect", http.StatusUnauthorized)
			return
		}

		session, _ := store.Get(r, "session")
		session.Values["username"] = username
		session.Save(r, w)

		http.Redirect(w, r, fmt.Sprintf("/user?username=%s", username), http.StatusSeeOther)
	}
}

func userHandler(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	if username == "" {
		http.Error(w, "Utilisateur non spécifié", http.StatusBadRequest)
		return
	}

	if r.Method == "GET" {
		var email, profilePicture string
		query := `SELECT email, profile_picture FROM users WHERE username = ?`
		err := db.QueryRowContext(context.Background(), query, username).Scan(&email, &profilePicture)
		if err != nil {
			http.Error(w, "Utilisateur non trouvé", http.StatusNotFound)
			return
		}

		data := struct {
			Username       string
			Email          string
			ProfilePicture string
		}{
			Username:       username,
			Email:          email,
			ProfilePicture: profilePicture,
		}

		tmpl, err := template.ParseFiles("templates/user.html")
		if err != nil {
			http.Error(w, "Erreur de lecture du fichier HTML", http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, data)
	} else if r.Method == "POST" {
		file, handler, err := r.FormFile("profile_picture")
		if err != nil {
			http.Error(w, "Erreur lors du téléchargement du fichier", http.StatusInternalServerError)
			return
		}
		defer file.Close()

		os.MkdirAll("static/uploads", os.ModePerm)

		filePath := filepath.Join("static/uploads", handler.Filename)
		f, err := os.Create(filePath)
		if err != nil {
			http.Error(w, "Erreur lors de la sauvegarde du fichier", http.StatusInternalServerError)
			return
		}
		defer f.Close()
		io.Copy(f, file)

		updateSQL := `UPDATE users SET profile_picture = ? WHERE username = ?`
		_, err = db.ExecContext(context.Background(), updateSQL, "/static/uploads/"+handler.Filename, username)
		if err != nil {
			http.Error(w, "Erreur lors de la mise à jour de la photo de profil", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/user?username=%s", username), http.StatusSeeOther)
	}
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	session.Options.MaxAge = -1
	session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func ajouterUtilisateur(username, email, motDePasse, profilePicture string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(motDePasse), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = db.ExecContext(context.Background(), `INSERT INTO users (username, email, mot_de_passe, profile_picture) VALUES (?, ?, ?, ?)`,
		username, email, hashedPassword, profilePicture)
	if err != nil {
		return err
	}
	return nil
}

func verifierUtilisateur(username, motDePasse string) error {
	var motDePasseDB string
	err := db.QueryRowContext(context.Background(), "SELECT mot_de_passe FROM users WHERE username = ?", username).Scan(&motDePasseDB)
	if err != nil {
		return err
	}

	err = bcrypt.CompareHashAndPassword([]byte(motDePasseDB), []byte(motDePasse))
	if err != nil {
		return fmt.Errorf("mot de passe incorrect")
	}
	return nil
}

func postsHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.QueryContext(context.Background(), "SELECT title, content FROM posts")
	if err != nil {
		http.Error(w, "Erreur lors de la récupération des messages", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	tmpl, err := template.ParseFiles("templates/post.html")
	if err != nil {
		http.Error(w, "Erreur de lecture du fichier HTML", http.StatusInternalServerError)
		return
	}

	type Post struct {
		Title   string
		Content string
	}

	var Posts []Post
	for rows.Next() {
		var post Post
		if err := rows.Scan(&post.Title, &post.Content); err != nil {
			http.Error(w, "Erreur lors de la lecture des messages", http.StatusInternalServerError)
			return
		}
		Posts = append(Posts, post)
	}

	data := struct {
		Post []Post
	}{
		Post: Posts,
	}

	for _, post := range Posts {
		fmt.Println(post.Title, post.Content)
	}

	tmpl.Execute(w, data)
}

func createPost(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		fmt.Println("PostHandler")
		session, _ := store.Get(r, "session")
		fmt.Println(session.Values)
		username, ok := session.Values["username"]
		fmt.Println(username)
		title, content := r.FormValue("title"), r.FormValue("content")
		fmt.Println(title, content)
		if !ok {
			http.Error(w, "Vous devez être connecté pour poster un message", http.StatusUnauthorized)
			return
		}
		fmt.Println(title, content)
		if r.Method == "POST" {
			_, err := db.ExecContext(context.Background(), "INSERT INTO posts (user, title, content) VALUES (?, ?, ?)", username, title, content)
			if err != nil {
				fmt.Println(err)
				http.Error(w, "Erreur lors de la publication du message", http.StatusInternalServerError)
				return
			}
		}
		fmt.Println("Post ajouté avec succès !")
		http.Redirect(w, r, "/post", http.StatusSeeOther)

	}
	http.ServeFile(w, r, "templates/create-post.html")
}


func topicsHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.QueryContext(context.Background(), "SELECT title, description FROM topics")
	if err != nil {
		http.Error(w, "Erreur lors de la récupération des messages", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	tmpl, err := template.ParseFiles("templates/topic.html")
	if err != nil {
		http.Error(w, "Erreur de lecture du fichier HTML", http.StatusInternalServerError)
		return
	}

	type Topic struct {
		Title       string
		Description string
	}

	var Topics []Topic
	for rows.Next() {
		var topic Topic
		if err := rows.Scan(&topic.Title, &topic.Description); err != nil {
			http.Error(w, "Erreur lors de la lecture des messages", http.StatusInternalServerError)
			return
		}
		Topics = append(Topics, topic)
	}

	data := struct {
		Topic []Topic
	}{
		Topic: Topics,
	}

	tmpl.Execute(w, data)
}

func createTopic(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {

		session, _ := store.Get(r, "session")

		username, ok := session.Values["username"]

		title, description := r.FormValue("title"), r.FormValue("description")

		if !ok {
			http.Error(w, "Vous devez être connecté pour poster un message", http.StatusUnauthorized)
			return
		}

		if r.Method == "POST" {
			_, err := db.ExecContext(context.Background(), "INSERT INTO topics (user, title, description) VALUES (?, ?, ?)", username, title, description)
			if err != nil {
				fmt.Println(err)
				http.Error(w, "Erreur lors de la publication du message", http.StatusInternalServerError)
				return
			}
		}
		fmt.Println("topic ajouté avec succès !")
		http.Redirect(w, r, "/topics", http.StatusSeeOther)

	}
	http.ServeFile(w, r, "templates/create-topic.html")
}
