package chatHeaven

import (
	"context"
	"html/template"
	"net/http"
	"time"

	_ "modernc.org/sqlite"
)

func TopicsHandler(w http.ResponseWriter, r *http.Request) {
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

	rows, err := db.QueryContext(context.Background(), "SELECT title, description FROM topics")
	if err != nil {
		http.Error(w, "Error retrieving the messages", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	tmpl, err := template.ParseFiles("templates/topic.html")
	if err != nil {
		http.Error(w, "Error reading the HTML file", http.StatusInternalServerError)
		return
	}

	var Topics []Topic
	for rows.Next() {
		var topic Topic
		if err := rows.Scan(&topic.Title, &topic.Description); err != nil {
			http.Error(w, "Error reading the messages", http.StatusInternalServerError)
			return
		}
		topic.NbPosts = len(getPosts(topic.Title))
		topic.LastPost = getLastPost(topic.Title)
		Topics = append(Topics, topic)
	}

	data := struct {
		Topic          []Topic
		Username       string
		ProfilePicture string
	}{
		Topic:          Topics,
		Username:       username.(string),
		ProfilePicture: profilePicture,
	}

	tmpl.Execute(w, data)
}

func CreateTopic(w http.ResponseWriter, r *http.Request) {
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

	tmpl, err := template.ParseFiles("templates/create-topic.html")
	if err != nil {
		http.Error(w, "Error reading the HTML file", http.StatusInternalServerError)
		return
	}

	if r.Method == "POST" {

		session, _ := store.Get(r, "session")

		username, ok := session.Values["username"]

		title, description := r.FormValue("title"), r.FormValue("description")

		date := time.Now().Format("02-01-2006 15:04")

		if !ok {
			http.Error(w, "You must be logged in to post a message", http.StatusUnauthorized)
			return
		}
		_, err := db.ExecContext(context.Background(), "INSERT INTO topics (user, title, description, date) VALUES (?, ?, ?, ?)", username, title, description, date)
		if err != nil {
			http.Error(w, "Error posting the message", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/topics", http.StatusSeeOther)

	}
	data := struct {
		Username       string
		ProfilePicture string
	}{
		Username:       username.(string),
		ProfilePicture: profilePicture,
	}

	tmpl.Execute(w, data)
}
