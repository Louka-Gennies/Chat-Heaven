package chatHeaven

import (
	"database/sql"
	"log"
	"os"

	"github.com/gorilla/sessions"
	_ "modernc.org/sqlite"
)

var db *sql.DB
var store = sessions.NewCookieStore([]byte("menu-classique-burger"))

func openDB() {
	dbPath := "./chatHeaven.db"

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		log.Fatal("Database file does not exist: ", dbPath)
	}

	var err error
	db, err = sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatal(err)
	}

	testQuery := "SELECT 1;"
	_, err = db.Query(testQuery)
	if err != nil {
		log.Fatal("Failed to execute test query: ", err)
	}

}
