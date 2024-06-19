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
		ErrorHandler(w, r)
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
			Last4Topics:    getTopics(4),
			IsLogged:       false,
		}

		tmpl, err := template.ParseFiles("templates/home.html")
		if err != nil {
			ErrorHandler(w, r)
			return
		}
		tmpl.Execute(w, data)
		return
	} else if ok {
		var profilePicture string
		err := db.QueryRowContext(context.Background(), "SELECT profile_picture FROM users WHERE username = ?", username).Scan(&profilePicture)
		if err != nil {
			ErrorHandler(w, r)
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
			Last4Topics:    getTopics(4),
			IsLogged:       true,
		}

		tmpl, err := template.ParseFiles("templates/home.html")
		if err != nil {
			ErrorHandler(w, r)
			return
		}
		tmpl.Execute(w, data)
	}
}
