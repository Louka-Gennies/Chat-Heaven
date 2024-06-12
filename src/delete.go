package chatHeaven

import (
	"context"
	"fmt"
	"net/http"

	_ "modernc.org/sqlite"
)

func DeletePost(w http.ResponseWriter, r *http.Request) {
	postTitle := r.URL.Query().Get("post")
	if postTitle == "" {
		http.Error(w, "Post not specified", http.StatusBadRequest)
		return
	}

	_, err := db.ExecContext(context.Background(), "DELETE FROM posts WHERE title = ?", postTitle)
	if err != nil {
		http.Error(w, "Error deleting the post", http.StatusInternalServerError)
		return
	}

	username := r.URL.Query().Get("user")

	http.Redirect(w, r, fmt.Sprintf("/user?username=%s", username), http.StatusSeeOther)
}

func DeleteTopic(w http.ResponseWriter, r *http.Request) {
	topicTitle := r.URL.Query().Get("topic")
	if topicTitle == "" {
		http.Error(w, "Topic not specified", http.StatusBadRequest)
		return
	}

	_, err := db.ExecContext(context.Background(), "DELETE FROM topics WHERE title = ?", topicTitle)
	if err != nil {
		http.Error(w, "Error deleting the topic", http.StatusInternalServerError)
		return
	}

	username := r.URL.Query().Get("user")

	http.Redirect(w, r, fmt.Sprintf("/user?username=%s", username), http.StatusSeeOther)
}