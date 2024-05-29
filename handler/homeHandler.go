package handler

import (
	"context"
	"html/template"
	"net/http"
)

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
		NbPosts        int
		NbTopics       int
		NbUsers        int
		Last4Topics    []Topic

	}{
		Username:       username.(string),
		ProfilePicture: profilePicture,
		NbPosts:        countPosts(),
		NbTopics:       countTopics(),
		NbUsers:        countUsers(),
		Last4Topics:    getTopics(4),
	}

	tmpl, err := template.ParseFiles("templates/home.html")
	if err != nil {
		http.Error(w, "Erreur de lecture du fichier HTML", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, data)
}