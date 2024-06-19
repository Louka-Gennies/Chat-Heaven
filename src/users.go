package chatHeaven

import (
	"context"
	"html/template"
	"net/http"
)

func UsersHandler(w http.ResponseWriter, r *http.Request) {
	openDB()
	session, _ := store.Get(r, "session")
	username, ok := session.Values["username"]
	if !ok {
		tmpl, err := template.ParseFiles("templates/home.html")
		if err != nil {
			http.Error(w, "Error reading the HTML file", http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, nil)
		return
	}

	var profilePicture string
	err := db.QueryRowContext(context.Background(), "SELECT profile_picture FROM users WHERE username = ?", username).Scan(&profilePicture)
	if err != nil {
		http.Error(w, "Error retrieving the profile picture", http.StatusInternalServerError)
		return
	}

	rows, err := db.QueryContext(context.Background(), "SELECT username, profile_picture FROM users")
	if err != nil {
		http.Error(w, "Error retrieving the messages", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	tmpl, err := template.ParseFiles("templates/users.html")
	if err != nil {
		http.Error(w, "Error reading the HTML file", http.StatusInternalServerError)
		return
	}

	var Users []User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.Username, &user.ProfilePicture); err != nil {
			http.Error(w, "Error reading the messages", http.StatusInternalServerError)
			return
		}
		user.LastPost = getLastPostFromUser(user.Username)
		user.TotalLikes = getTotalLikesFromUser(user.Username)
		Users = append(Users, user)
	}

	data := struct {
		Users          []User
		Username       string
		ProfilePicture string
	}{
		Users,
		username.(string),
		profilePicture,
	}

	tmpl.Execute(w, data)
}
