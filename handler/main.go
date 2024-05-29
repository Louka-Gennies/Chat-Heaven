package handler

import (
	"database/sql"
	"github.com/gorilla/sessions"
)

var db *sql.DB
var store = sessions.NewCookieStore([]byte("menu-classique-burger"))

type Topic struct {
	Title       string
	Description string
	NbPosts	 int
}

type Post struct {
	Title   string
	Content string
	User   string
	Topic  string
}