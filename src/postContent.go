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
		ErrorHandler(w, r)
		return
	}
	session, _ := store.Get(r, "session")
	username, ok := session.Values["username"]
	if !ok {
		var title, content, user, topic, picture string
		err = db.QueryRowContext(context.Background(), "SELECT title, content, user, topic, picture FROM posts WHERE id = ?", postIDInt).Scan(&title, &content, &user, &topic, &picture)

		if err != nil {
			ErrorHandler(w, r)
			return
		}

		if r.Method == "POST" {
			session, _ := store.Get(r, "session")
			username, ok := session.Values["username"]
			comment := r.FormValue("content")
			date := time.Now().Format("02-01-2006 15:04")
			if !ok {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}
			_, err := db.ExecContext(context.Background(), "INSERT INTO comments (user, content, post, date) VALUES (?, ?, ?, ?)", username, comment, title, date)
			if err != nil {
				ErrorHandler(w, r)
				return
			}
			http.Redirect(w, r, fmt.Sprintf("/post-content?postID=%d", postIDInt), http.StatusSeeOther)

		}

		data := struct {
			Post           Post
			Comment        []Comment
			Username       interface{}
			ProfilePicture interface{}
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
				AlreadyLiked:          false,
				AlreadyDisliked:       false,
				Date:                  getDatePost(postIDInt),
				Picture:               picture,
			},
			Comment:        getComment(title),
			Username:       nil,
			ProfilePicture: nil,
		}

		tmpl, err := template.ParseFiles("templates/post-content.html")
		if err != nil {
			ErrorHandler(w, r)
			return
		}
		tmpl.Execute(w, data)
	} else {

		var profilePicture string
		err = db.QueryRowContext(context.Background(), "SELECT profile_picture FROM users WHERE username = ?", username).Scan(&profilePicture)
		if err != nil {
			ErrorHandler(w, r)
			return
		}
		if postID == "" {
			ErrorHandler(w, r)
			return
		}

		var title, content, user, topic, picture string
		err = db.QueryRowContext(context.Background(), "SELECT title, content, user, topic, picture FROM posts WHERE id = ?", postIDInt).Scan(&title, &content, &user, &topic, &picture)

		if err != nil {
			ErrorHandler(w, r)
			return
		}

		if r.Method == "POST" {
			session, _ := store.Get(r, "session")
			username, ok := session.Values["username"]
			comment := r.FormValue("content")
			date := time.Now().Format("02-01-2006 15:04")
			if !ok {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}
			_, err := db.ExecContext(context.Background(), "INSERT INTO comments (user, content, post, date) VALUES (?, ?, ?, ?)", username, comment, title, date)
			if err != nil {
				ErrorHandler(w, r)
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

		tmpl, err := template.ParseFiles("templates/post-content.html")
		if err != nil {
			ErrorHandler(w, r)
			return
		}
		tmpl.Execute(w, data)
	}
}
