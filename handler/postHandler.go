package handler

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
)

func postsHandler(w http.ResponseWriter, r *http.Request) {
	topic := r.URL.Query().Get("topic")
	fmt.Println(topic)
	Posts := getPosts(topic)

	for _, post := range Posts {
		fmt.Println(post.Title, post.Content)
	}

	tmpl, err := template.ParseFiles("templates/post.html")
	if err != nil {
		http.Error(w, "Erreur de lecture du fichier HTML", http.StatusInternalServerError)
		return
	}

	data := struct {
		Topic string
		Post []Post
	}{
		Topic: topic,
		Post: Posts,
	}

	for _, post := range Posts {
		fmt.Println(post.Title, post.Content)
	}


	tmpl.Execute(w, data)
}

func createPost(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/create-post.html")
	if err != nil {
		http.Error(w, "Erreur de lecture du fichier HTML", http.StatusInternalServerError)
		return
	}

	topic := r.URL.Query().Get("topic")

	data := struct {
		Topic string
	}{
		Topic: topic,
	}

	if r.Method == "POST" {
		session, _ := store.Get(r, "session")
		username, ok := session.Values["username"]
		title, content, topicTitle := r.FormValue("title"), r.FormValue("content"), r.FormValue("topic")

		if !ok {
			http.Error(w, "Vous devez être connecté pour poster un message", http.StatusUnauthorized)
			return
		}
		_, err := db.ExecContext(context.Background(), "INSERT INTO posts (user, title, content, topic) VALUES (?, ?, ?, ?)", username, title, content, topicTitle)
		if err != nil {
			fmt.Println(err)
			http.Error(w, "Erreur lors de la publication du message", http.StatusInternalServerError)
			return
		}
		fmt.Println("Post ajouté avec succès !")
		http.Redirect(w, r, "/post", http.StatusSeeOther)

	}

	
	tmpl.Execute(w, data)
}