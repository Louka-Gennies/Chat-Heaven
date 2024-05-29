package handler

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
)

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

	for _, topic := range Topics {
		fmt.Println(topic.Title, topic.Description, topic.NbPosts)
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