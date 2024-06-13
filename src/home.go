package chatHeaven

import (
	"context"
	"html/template"
	"net/http"

	_ "modernc.org/sqlite"
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	openDB()
	session, err := store.Get(r, "session")
	if err != nil {
		http.Error(w, "Error retrieving session", http.StatusInternalServerError)
		return
	}

	username, ok := session.Values["username"]

	if !ok {
		data := struct {
			Username       string
			ProfilePicture string
			NbPosts        int
			NbTopics       int
			NbUsers        int
			Last4Topics    []Topic
			IsLogged       bool
		}{
			Username:       "",
			ProfilePicture: "",
			NbPosts:        countPosts(),
			NbTopics:       countTopics(),
			NbUsers:        countUsers(),
			Last4Topics:    getTopics(3),
			IsLogged:       false,
		}

		tmpl, err := template.ParseFiles("templates/home.html")
		if err != nil {
			http.Error(w, "Error reading the HTML file", http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, data)
		return
	} else if ok {
		var profilePicture string
		err := db.QueryRowContext(context.Background(), "SELECT profile_picture FROM users WHERE username = ?", username).Scan(&profilePicture)
		if err != nil {
			http.Error(w, "Error retrieving the profile picture", http.StatusInternalServerError)
			return
		}

		data := struct {
			Username       string
			ProfilePicture string
			NbPosts        int
			NbTopics       int
			NbUsers        int
			Last4Topics    []Topic
			IsLogged       bool
		}{
			Username:       username.(string),
			ProfilePicture: profilePicture,
			NbPosts:        countPosts(),
			NbTopics:       countTopics(),
			NbUsers:        countUsers(),
			Last4Topics:    getTopics(3),
			IsLogged:       true,
		}

		tmpl, err := template.ParseFiles("templates/home.html")
		if err != nil {
			http.Error(w, "Error reading the HTML file", http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, data)
	}
}
