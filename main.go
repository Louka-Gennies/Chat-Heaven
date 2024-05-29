package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"

	"github.com/gorilla/sessions"
	_ "modernc.org/sqlite"

	"github.com/Louka-Gennies/Chat-Heaven/handler/handler"
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

func main() {
	handler.test()
	dbPath := "chatHeaven.db"

	var err error
	db, err = sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.ExecContext(context.Background(), `CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        username TEXT NOT NULL UNIQUE,
        email TEXT NOT NULL UNIQUE,
        mot_de_passe TEXT NOT NULL,
        profile_picture TEXT,
        user_likes INTEGER
    )`)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.ExecContext(context.Background(), `CREATE TABLE IF NOT EXISTS topics (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL UNIQUE,
		description TEXT NOT NULL,
		picture TEXT,
		user TEXT NOT NULL,
		topic_likes INTEGER,
		FOREIGN KEY (user) REFERENCES users(username)
	)`)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.ExecContext(context.Background(), `CREATE TABLE IF NOT EXISTS posts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL UNIQUE,
		content TEXT NOT NULL UNIQUE,
		picture TEXT,
		user TEXT NOT NULL,
		topic TEXT NOT NULL,
		post_likes INTEGER,
		FOREIGN KEY (user) REFERENCES users(username)
		FOREIGN KEY (topic) REFERENCES topics(title)
	)`)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", handler.homeHandler)
	http.HandleFunc("/register", handler.registerHandler)
	http.HandleFunc("/login", handler.loginHandler)
	http.HandleFunc("/user", handler.userHandler)
	http.HandleFunc("/logout", handler.logoutHandler)
	http.HandleFunc("/create-post", handler.createPost)
	http.HandleFunc("/posts", handler.postsHandler)
	http.HandleFunc("/topics", handler.topicsHandler)
	http.HandleFunc("/create-topic", handler.createTopic)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	log.Println("Server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
