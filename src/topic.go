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
		rows, err := db.QueryContext(context.Background(), "SELECT title, description FROM topics")
		if err != nil {
			ErrorHandler(w, r)
			return
		}
		defer rows.Close()

		tmpl, err := template.ParseFiles("templates/topic.html")
		if err != nil {
			ErrorHandler(w, r)
			return
		}

		var Topics []Topic
		for rows.Next() {
			var topic Topic
			if err := rows.Scan(&topic.Title, &topic.Description); err != nil {
				ErrorHandler(w, r)
				return
			}
			topic.NbPosts = len(getPosts(topic.Title, ""))
			topic.LastPost = getLastPost(topic.Title)
			Topics = append(Topics, topic)
		}

		data := struct {
			Topic          []Topic
			Username       interface{}
			ProfilePicture interface{}
		}{
			Topic:          Topics,
			Username:       nil,
			ProfilePicture: nil,
		}

		tmpl.Execute(w, data)

	} else {
		var profilePicture string
		err := db.QueryRowContext(context.Background(), "SELECT profile_picture FROM users WHERE username = ?", username).Scan(&profilePicture)
		if err != nil {
			ErrorHandler(w, r)
			return
		}

		rows, err := db.QueryContext(context.Background(), "SELECT title, description FROM topics")
		if err != nil {
			ErrorHandler(w, r)
			return
		}
		defer rows.Close()

		tmpl, err := template.ParseFiles("templates/topic.html")
		if err != nil {
			ErrorHandler(w, r)
			return
		}

		var Topics []Topic
		for rows.Next() {
			var topic Topic
			if err := rows.Scan(&topic.Title, &topic.Description); err != nil {
				ErrorHandler(w, r)
				return
			}
			topic.NbPosts = len(getPosts(topic.Title, ""))
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
}

func CreateTopic(w http.ResponseWriter, r *http.Request) {
	openDB()
	session, _ := store.Get(r, "session")
	username, ok := session.Values["username"]
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	var profilePicture string
	err := db.QueryRowContext(context.Background(), "SELECT profile_picture FROM users WHERE username = ?", username).Scan(&profilePicture)
	if err != nil {
		ErrorHandler(w, r)
		return
	}

	tmpl, err := template.ParseFiles("templates/create-topic.html")
	if err != nil {
		ErrorHandler(w, r)
		return
	}

	if r.Method == "POST" {

		session, _ := store.Get(r, "session")

		username, ok := session.Values["username"]

		title, description := r.FormValue("title"), r.FormValue("description")

		date := time.Now().Format("02-01-2006 15:04")

		if !ok {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		_, err := db.ExecContext(context.Background(), "INSERT INTO topics (user, title, description, date) VALUES (?, ?, ?, ?)", username, title, description, date)
		if err != nil {
			ErrorHandler(w, r)
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
