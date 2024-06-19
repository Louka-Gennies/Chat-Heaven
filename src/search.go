package chatHeaven

import (
	"context"
	"encoding/json"
	"net/http"

	_ "modernc.org/sqlite"
)

func SearchAutocomplete(w http.ResponseWriter, r *http.Request) {
	openDB()
	search := r.URL.Query().Get("search")

	rows, err := db.QueryContext(context.Background(), "SELECT username, profile_picture FROM users WHERE username LIKE ?", search+"%")
	if err != nil {
		ErrorHandler(w, r)
		return
	}
	defer rows.Close()

	var results []map[string]string
	for rows.Next() {
		var user, profilePicture string
		if err := rows.Scan(&user, &profilePicture); err != nil {
			ErrorHandler(w, r)
			return
		}
		results = append(results, map[string]string{"type": "user", "value": user, "profil_picture": profilePicture})
	}

	rows, err = db.QueryContext(context.Background(), "SELECT title FROM topics WHERE title LIKE ?", search+"%")
	if err != nil {
		ErrorHandler(w, r)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var topic string
		if err := rows.Scan(&topic); err != nil {
			ErrorHandler(w, r)
			return
		}
		results = append(results, map[string]string{"type": "topic", "value": topic})
	}

	jsonData, err := json.Marshal(results)
	if err != nil {
		ErrorHandler(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}
