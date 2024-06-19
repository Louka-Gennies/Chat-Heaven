package chatHeaven

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"time"

	_ "modernc.org/sqlite"
)

func PostsHandler(w http.ResponseWriter, r *http.Request) {
	openDB()
	topic := r.URL.Query().Get("topic")
	sort := r.URL.Query().Get("sort")
	Posts := getPosts(topic, sort)
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

	tmpl, err := template.ParseFiles("templates/post.html")
	if err != nil {
		http.Error(w, "Error reading the HTML file", http.StatusInternalServerError)
		return
	}

	data := struct {
		Topic            string
		TopicDescription string
		Post             []Post
		Username         string
		ProfilePicture   string
	}{
		Topic:            topic,
		TopicDescription: getTopicDescription(topic),
		Post:             Posts,
		Username:         username.(string),
		ProfilePicture:   profilePicture,
	}

	tmpl.Execute(w, data)
}

func CreatePost(w http.ResponseWriter, r *http.Request) {
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

	tmpl, err := template.ParseFiles("templates/create-post.html")
	if err != nil {
		http.Error(w, "Error reading the HTML file", http.StatusInternalServerError)
		return
	}

	topic := r.URL.Query().Get("topic")

	data := struct {
		Topic          string
		Username       string
		ProfilePicture string
	}{
		Topic:          topic,
		Username:       username.(string),
		ProfilePicture: profilePicture,
	}

	if r.Method == "POST" {
		session, _ := store.Get(r, "session")
		username, ok := session.Values["username"]
		title, content, topicTitle := r.FormValue("title"), r.FormValue("content"), r.FormValue("topic")
		date := time.Now().Format("02-01-2006 15:04")

		file, handler, err := r.FormFile("picture")
		if err != nil {
			http.Error(w, "Error during file upload", http.StatusInternalServerError)
			return
		}
		defer file.Close()

		os.MkdirAll("uploads", os.ModePerm)

		filePath := "uploads/" + handler.Filename
		f, err := os.Create(filePath)
		if err != nil {
			http.Error(w, "Error saving the file", http.StatusInternalServerError)
			return
		}
		defer f.Close()
		io.Copy(f, file)

		if !ok {
			http.Error(w, "You must be logged in to post a message", http.StatusUnauthorized)
			return
		}
		_, err = db.ExecContext(context.Background(), "INSERT INTO posts (user, title, content, picture, topic, date) VALUES (?, ?, ?, ?, ?, ?)", username, title, content, filePath, topicTitle, date)
		if err != nil {
			http.Error(w, "Error posting the message", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, fmt.Sprintf("/posts?topic=%s", topic), http.StatusSeeOther)

	}

	tmpl.Execute(w, data)
}
