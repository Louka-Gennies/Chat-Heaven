package chatHeaven

import (
	"context"
	"html/template"
	"net/http"
)

func UsersHandler(w http.ResponseWriter, r *http.Request) {
	openDB()
	session, _ := store.Get(r, "session")
	username, ok := session.Values["username"]
	if !ok {
		rows, err := db.QueryContext(context.Background(), "SELECT username, profile_picture FROM users")
		if err != nil {
			ErrorHandler(w, r)
			return
		}
		defer rows.Close()

		tmpl, err := template.ParseFiles("templates/users.html")
		if err != nil {
			ErrorHandler(w, r)
			return
		}

		var Users []User
		for rows.Next() {
			var user User
			if err := rows.Scan(&user.Username, &user.ProfilePicture); err != nil {
				ErrorHandler(w, r)
				return
			}
			user.LastPost = getLastPostFromUser(user.Username)
			user.TotalLikes = getTotalLikesFromUser(user.Username)
			Users = append(Users, user)
		}

		data := struct {
			Users          []User
			Username       interface{}
			ProfilePicture interface{}
		}{
			Users,
			nil,
			nil,
		}

		tmpl.Execute(w, data)
	} else {

		var profilePicture string
		err := db.QueryRowContext(context.Background(), "SELECT profile_picture FROM users WHERE username = ?", username).Scan(&profilePicture)
		if err != nil {
			ErrorHandler(w, r)
			return
		}

		rows, err := db.QueryContext(context.Background(), "SELECT username, profile_picture FROM users")
		if err != nil {
			ErrorHandler(w, r)
			return
		}
		defer rows.Close()

		tmpl, err := template.ParseFiles("templates/users.html")
		if err != nil {
			ErrorHandler(w, r)
			return
		}

		var Users []User
		for rows.Next() {
			var user User
			if err := rows.Scan(&user.Username, &user.ProfilePicture); err != nil {
				ErrorHandler(w, r)
				return
			}
			user.LastPost = getLastPostFromUser(user.Username)
			user.TotalLikes = getTotalLikesFromUser(user.Username)
			Users = append(Users, user)
		}

		data := struct {
			Users          []User
			Username       string
			ProfilePicture string
		}{
			Users,
			username.(string),
			profilePicture,
		}

		tmpl.Execute(w, data)
	}
}
