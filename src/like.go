package chatHeaven

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	_ "modernc.org/sqlite"
)

func AddLike(w http.ResponseWriter, r *http.Request) {
	postID := r.URL.Query().Get("postID")
	postIDInt, err := strconv.Atoi(postID)
	if err != nil {
		fmt.Println(err)
	}
	if postID == "" {
		http.Error(w, "Message not specified", http.StatusBadRequest)
		return
	}

	session, _ := store.Get(r, "session")
	username, ok := session.Values["username"]
	if !ok {
		http.Error(w, "You must be logged in to like a message", http.StatusUnauthorized)
		return
	}

	var existingLike int
	var existingDislike int
	err = db.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM likes WHERE user = ? AND title = ?", username, postIDInt).Scan(&existingLike)
	if err != nil {
		http.Error(w, "Error checking the likes", http.StatusInternalServerError)
		return
	}
	err = db.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM dislikes WHERE user = ? AND title = ?", username, postIDInt).Scan(&existingDislike)
	if err != nil {
		http.Error(w, "Error checking the dislikes", http.StatusInternalServerError)
		return
	}

	if existingLike > 0 {
		// If a like exists, delete it
		_, err = db.ExecContext(context.Background(), "DELETE FROM likes WHERE user = ? AND title = ?", username, postIDInt)
		if err != nil {
			http.Error(w, "Error deleting the like", http.StatusInternalServerError)
			return
		}
	} else if existingDislike > 0 {
		// If no like exists, add a new one
		_, err = db.ExecContext(context.Background(), "INSERT INTO likes (user, title) VALUES (?, ?)", username, postIDInt)
		if err != nil {
			http.Error(w, "Error adding the like", http.StatusInternalServerError)
			return
		}
		_, err = db.ExecContext(context.Background(), "DELETE FROM dislikes WHERE user = ? AND title = ?", username, postIDInt)
		if err != nil {
			http.Error(w, "Error adding the dislike", http.StatusInternalServerError)
			return
		}
	} else {
		_, err = db.ExecContext(context.Background(), "INSERT INTO likes (user, title) VALUES (?, ?)", username, postIDInt)
		if err != nil {
			http.Error(w, "Error adding the like", http.StatusInternalServerError)
			return
		}
	}

	http.Redirect(w, r, fmt.Sprintf("/post-content?postID=%d", postIDInt), http.StatusSeeOther)
}

func AddDislike(w http.ResponseWriter, r *http.Request) {
	postID := r.URL.Query().Get("postID")
	postIDInt, err := strconv.Atoi(postID)
	if err != nil {
		fmt.Println(err)
	}
	if postID == "" {
		http.Error(w, "Message not specified", http.StatusBadRequest)
		return
	}

	session, _ := store.Get(r, "session")
	username, ok := session.Values["username"]
	if !ok {
		http.Error(w, "You must be logged in to like a message", http.StatusUnauthorized)
		return
	}

	var existingDislike int
	var existingLike int
	err = db.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM dislikes WHERE user = ? AND title = ?", username, postIDInt).Scan(&existingDislike)
	if err != nil {
		http.Error(w, "Error checking the dislikes", http.StatusInternalServerError)
		return
	}
	err = db.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM likes WHERE user = ? AND title = ?", username, postIDInt).Scan(&existingLike)
	if err != nil {
		http.Error(w, "Error checking the likes", http.StatusInternalServerError)
		return
	}

	if existingDislike > 0 {
		// If a dislike exists, delete it
		_, err = db.ExecContext(context.Background(), "DELETE FROM dislikes WHERE user = ? AND title = ?", username, postIDInt)
		if err != nil {
			http.Error(w, "Error deleting the dislike", http.StatusInternalServerError)
			return
		}
	} else if existingLike > 0 {
		// If no dislike exists, add a new one
		_, err = db.ExecContext(context.Background(), "INSERT INTO dislikes (user, title) VALUES (?, ?)", username, postIDInt)
		if err != nil {
			http.Error(w, "Error adding the dislike", http.StatusInternalServerError)
			return
		}
		_, err = db.ExecContext(context.Background(), "DELETE FROM likes WHERE user = ? AND title = ?", username, postIDInt)
		if err != nil {
			http.Error(w, "Error adding the like", http.StatusInternalServerError)
			return
		}
	} else {
		_, err = db.ExecContext(context.Background(), "INSERT INTO dislikes (user, title) VALUES (?, ?)", username, postIDInt)
		if err != nil {
			http.Error(w, "Error adding the dislike", http.StatusInternalServerError)
			return
		}
	}

	http.Redirect(w, r, fmt.Sprintf("/post-content?postID=%d", postIDInt), http.StatusSeeOther)
}

func isLiked(username string, postID int) bool {
	var existingLike int
	err := db.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM likes WHERE user = ? AND title = ?", username, postID).Scan(&existingLike)
	if err != nil {
		return false
	}
	return existingLike > 0
}

func isDisliked(username string, postID int) bool {
	var existingDislike int
	err := db.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM dislikes WHERE user = ? AND title = ?", username, postID).Scan(&existingDislike)
	if err != nil {
		return false
	}
	return existingDislike > 0
}