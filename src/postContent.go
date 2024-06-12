package chatHeaven

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"time"

	_ "modernc.org/sqlite"
)

func GetPostContent(w http.ResponseWriter, r *http.Request) {
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
	fmt.Println(err)
	if postID == "" {
		http.Error(w, "Message not specified", http.StatusBadRequest)
		return
	}

	var title, content, user, topic, picture string
	err = db.QueryRowContext(context.Background(), "SELECT title, content, user, topic, picture FROM posts WHERE id = ?", postIDInt).Scan(&title, &content, &user, &topic, &picture)

	if err != nil {
		http.Error(w, "Message not found", http.StatusNotFound)
		return
	}

	if r.Method == "POST" {
		session, _ := store.Get(r, "session")
		username, ok := session.Values["username"]
		comment := r.FormValue("content")
		date := time.Now().Format("02-01-2006 15:04")
		if !ok {
			http.Error(w, "You must be logged in to comment on a message", http.StatusUnauthorized)
			return
		}
		_, err := db.ExecContext(context.Background(), "INSERT INTO comments (user, content, post, date) VALUES (?, ?, ?, ?)", username, comment, title, date)
		if err != nil {
			http.Error(w, "Error posting the comment", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, fmt.Sprintf("/post-content?postID=%d", postIDInt), http.StatusSeeOther)

	}

	data := struct {
		Post           Post
		Comment        []Comment
		Username       string
		ProfilePicture string
	}{
		Post: Post{
			ID:                    postIDInt,
			Title:                 title,
			Content:               content,
			User:                  user,
			Topic:                 topic,
			Likes:                 likeCount(postIDInt),
			Dislikes:              dislikeCount(postIDInt),
			LikeDislikeDifference: likeCount(postIDInt) - dislikeCount(postIDInt),
			AlreadyLiked:          isLiked(username.(string), postIDInt),
			AlreadyDisliked:       isDisliked(username.(string), postIDInt),
			Date:                  getDatePost(postIDInt),
			Picture:               picture,
		},
		Comment:        getComment(title),
		Username:       username.(string),
		ProfilePicture: profilePicture,
	}

	fmt.Print(data.Post.Picture)

	tmpl, err := template.ParseFiles("templates/post-content.html")
	if err != nil {
		http.Error(w, "Error reading the HTML file", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, data)
}