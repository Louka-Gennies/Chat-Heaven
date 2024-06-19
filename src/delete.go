package chatHeaven

import (
	"context"
	"fmt"
	"net/http"

	_ "modernc.org/sqlite"
)

func DeletePost(w http.ResponseWriter, r *http.Request) {
	openDB()
	postTitle := r.URL.Query().Get("post")
	if postTitle == "" {
		ErrorHandler(w, r)
		return
	}

	_, err := db.ExecContext(context.Background(), "DELETE FROM posts WHERE title = ?", postTitle)
	if err != nil {
		ErrorHandler(w, r)
		return
	}

	username := r.URL.Query().Get("user")

	http.Redirect(w, r, fmt.Sprintf("/user?username=%s", username), http.StatusSeeOther)
}

func DeleteTopic(w http.ResponseWriter, r *http.Request) {
	openDB()
	topicTitle := r.URL.Query().Get("topic")
	if topicTitle == "" {
		ErrorHandler(w, r)
		return
	}

	_, err := db.ExecContext(context.Background(), "DELETE FROM topics WHERE title = ?", topicTitle)
	if err != nil {
		ErrorHandler(w, r)
		return
	}

	username := r.URL.Query().Get("user")

	http.Redirect(w, r, fmt.Sprintf("/user?username=%s", username), http.StatusSeeOther)
}

func DeleteComment(w http.ResponseWriter, r *http.Request) {
	openDB()
	commentID := r.URL.Query().Get("commentID")
	if commentID == "" {
		ErrorHandler(w, r)
		return
	}

	_, err := db.ExecContext(context.Background(), "DELETE FROM comments WHERE id = ?", commentID)
	if err != nil {
		ErrorHandler(w, r)
		return
	}

	username := r.URL.Query().Get("user")

	http.Redirect(w, r, fmt.Sprintf("/user?username=%s", username), http.StatusSeeOther)
}
