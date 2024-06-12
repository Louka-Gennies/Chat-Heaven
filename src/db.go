package chatHeaven

import (
	"database/sql"
	"log"

	"github.com/gorilla/sessions"
	_ "modernc.org/sqlite"
)

var db *sql.DB
var store = sessions.NewCookieStore([]byte("menu-classique-burger"))

func openDB() {
	dbPath := "../chatHeaven.db"

	var err error
	db, err = sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatal(err)
	}
}