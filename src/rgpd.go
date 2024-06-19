package chatHeaven

import (
	"context"
	"html/template"
	"net/http"
)

func RgpdHandler(w http.ResponseWriter, r *http.Request) {
	openDB()
	session, _ := store.Get(r, "session")
	username, ok := session.Values["username"]
	if !ok {
		data := struct {
			ProfilePicture interface{}
			Username       interface{}
		}{
			ProfilePicture: nil,
			Username:       nil,
		}
		tmpl, err := template.ParseFiles("templates/rgpd.html")
		if err != nil {
			ErrorHandler(w, r)
			return
		}
		tmpl.Execute(w, data)
	} else {
		var profilePicture string
		err := db.QueryRowContext(context.Background(), "SELECT profile_picture FROM users WHERE username = ?", username).Scan(&profilePicture)
		if err != nil {
			ErrorHandler(w, r)
			return
		}
		data := struct {
			ProfilePicture string
			Username       string
		}{
			ProfilePicture: profilePicture,
			Username:       username.(string),
		}
		tmpl, err := template.ParseFiles("templates/rgpd.html")
		if err != nil {
			ErrorHandler(w, r)
			return
		}
		tmpl.Execute(w, data)
	}
}
