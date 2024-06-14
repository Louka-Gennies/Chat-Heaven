package chatHeaven

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
)

func EditPost(w http.ResponseWriter, r *http.Request) {
	openDB()
	postID := r.URL.Query().Get("postID")
	postIDInt, err := strconv.Atoi(postID)
	if err != nil {
		http.Error(w, "Error converting the post ID", http.StatusInternalServerError)
		return
	}
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
	err = db.QueryRowContext(context.Background(), "SELECT profile_picture FROM users WHERE username = ?", username).Scan(&profilePicture)
	if err != nil {
		http.Error(w, "Error retrieving the profile picture", http.StatusInternalServerError)
		return
	}

	var title, content, user, topic, picture string
	err = db.QueryRowContext(context.Background(), "SELECT title, content, user, topic, picture FROM posts WHERE id = ?", postIDInt).Scan(&title, &content, &user, &topic, &picture)
	if err != nil {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	if r.Method == "POST" {
		newTitle := r.FormValue("title")
		newContent := r.FormValue("content")

		_, err := db.ExecContext(context.Background(), "UPDATE posts SET title = ?, content = ? WHERE id = ?", newTitle, newContent, postIDInt)
		if err != nil {
			http.Error(w, "Error updating the post", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, fmt.Sprintf("/post-content?postID=%d", postIDInt), http.StatusSeeOther)
	} else {
		data := struct {
			Post           Post
			Username       string
			ProfilePicture string
		}{
			Post: Post{
				ID:      postIDInt,
				Title:   title,
				Content: content,
				User:    user,
				Topic:   topic,
				Picture: picture,
			},
			Username:       username.(string),
			ProfilePicture: profilePicture,
		}

		tmpl, err := template.ParseFiles("templates/edit-post.html")
		if err != nil {
			http.Error(w, "Error reading the HTML file", http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, data)
	}
}
