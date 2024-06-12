package chatHeaven

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	_ "modernc.org/sqlite"
)

func SearchAutocomplete(w http.ResponseWriter, r *http.Request) {
	openDB()
	search := r.URL.Query().Get("search")

	rows, err := db.QueryContext(context.Background(), "SELECT username, profile_picture FROM users WHERE username LIKE ?", search+"%")
	if err != nil {
		http.Error(w, "Error retrieving the users", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var results []map[string]string
	for rows.Next() {
		var user, profilePicture string
		if err := rows.Scan(&user, &profilePicture); err != nil {
			http.Error(w, "Error reading the users", http.StatusInternalServerError)
			return
		}
		results = append(results, map[string]string{"type": "user", "value": user, "profil_picture": profilePicture})
	}

	rows, err = db.QueryContext(context.Background(), "SELECT title FROM topics WHERE title LIKE ?", search+"%")
	if err != nil {
		http.Error(w, "Error retrieving the topics", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var topic string
		if err := rows.Scan(&topic); err != nil {
			http.Error(w, "Error reading the topics", http.StatusInternalServerError)
			return
		}
		results = append(results, map[string]string{"type": "topic", "value": topic})
	}
	fmt.Println(results)

	jsonData, err := json.Marshal(results)
	if err != nil {
		http.Error(w, "Error converting results to JSON", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}