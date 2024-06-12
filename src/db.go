package chatHeaven

import (
    "database/sql"
    "fmt"
    "log"
    "os"

    "github.com/gorilla/sessions"
    _ "modernc.org/sqlite"
)

var db *sql.DB
var store = sessions.NewCookieStore([]byte("menu-classique-burger"))

func openDB() {
    dbPath := "../chatHeaven.db"

    if _, err := os.Stat(dbPath); os.IsNotExist(err) {
        log.Fatal("Database file does not exist: ", dbPath)
    }

    fmt.Println("Opening database")

    var err error
    db, err = sql.Open("sqlite", dbPath)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("Testing database connection")

    testQuery := "SELECT 1;"
    _, err = db.Query(testQuery)
    if err != nil {
        log.Fatal("Failed to execute test query: ", err)
    }

    fmt.Println("Content of the database")

	rows, err := db.Query("SELECT name FROM sqlite_master WHERE type='table';")
	if err != nil {
		log.Fatal("Failed to retrieve table names: ", err)
	}
	defer rows.Close()
	
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			log.Fatal("Failed to scan table name: ", err)
		}
		fmt.Println("Table name: ", name)
	}
}
