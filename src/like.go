package chatHeaven

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	_ "modernc.org/sqlite"
)

func AddLike(w http.ResponseWriter, r *http.Request) {
	openDB()
	postID := r.URL.Query().Get("postID")
	postIDInt, err := strconv.Atoi(postID)
	if err != nil {
	}
	if postID == "" {
		ErrorHandler(w, r)
		return
	}

	session, _ := store.Get(r, "session")
	username, ok := session.Values["username"]
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	var existingLike int
	var existingDislike int
	err = db.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM likes WHERE user = ? AND title = ?", username, postIDInt).Scan(&existingLike)
	if err != nil {
		ErrorHandler(w, r)
		return
	}
	err = db.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM dislikes WHERE user = ? AND title = ?", username, postIDInt).Scan(&existingDislike)
	if err != nil {
		ErrorHandler(w, r)
		return
	}

	if existingLike > 0 {
		_, err = db.ExecContext(context.Background(), "DELETE FROM likes WHERE user = ? AND title = ?", username, postIDInt)
		if err != nil {
			ErrorHandler(w, r)
			return
		}
	} else if existingDislike > 0 {
		_, err = db.ExecContext(context.Background(), "INSERT INTO likes (user, title) VALUES (?, ?)", username, postIDInt)
		if err != nil {
			ErrorHandler(w, r)
			return
		}
		_, err = db.ExecContext(context.Background(), "DELETE FROM dislikes WHERE user = ? AND title = ?", username, postIDInt)
		if err != nil {
			ErrorHandler(w, r)
			return
		}
	} else {
		_, err = db.ExecContext(context.Background(), "INSERT INTO likes (user, title) VALUES (?, ?)", username, postIDInt)
		if err != nil {
			ErrorHandler(w, r)
			return
		}
	}

	http.Redirect(w, r, fmt.Sprintf("/post-content?postID=%d", postIDInt), http.StatusSeeOther)
}

func AddDislike(w http.ResponseWriter, r *http.Request) {
	openDB()
	postID := r.URL.Query().Get("postID")
	postIDInt, err := strconv.Atoi(postID)
	if err != nil {
	}
	if postID == "" {
		ErrorHandler(w, r)
		return
	}

	session, _ := store.Get(r, "session")
	username, ok := session.Values["username"]
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	var existingDislike int
	var existingLike int
	err = db.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM dislikes WHERE user = ? AND title = ?", username, postIDInt).Scan(&existingDislike)
	if err != nil {
		ErrorHandler(w, r)
		return
	}
	err = db.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM likes WHERE user = ? AND title = ?", username, postIDInt).Scan(&existingLike)
	if err != nil {
		ErrorHandler(w, r)
		return
	}

	if existingDislike > 0 {
		_, err = db.ExecContext(context.Background(), "DELETE FROM dislikes WHERE user = ? AND title = ?", username, postIDInt)
		if err != nil {
			ErrorHandler(w, r)
			return
		}
	} else if existingLike > 0 {
		_, err = db.ExecContext(context.Background(), "INSERT INTO dislikes (user, title) VALUES (?, ?)", username, postIDInt)
		if err != nil {
			ErrorHandler(w, r)
			return
		}
		_, err = db.ExecContext(context.Background(), "DELETE FROM likes WHERE user = ? AND title = ?", username, postIDInt)
		if err != nil {
			ErrorHandler(w, r)
			return
		}
	} else {
		_, err = db.ExecContext(context.Background(), "INSERT INTO dislikes (user, title) VALUES (?, ?)", username, postIDInt)
		if err != nil {
			ErrorHandler(w, r)
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
