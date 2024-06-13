package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"

	_ "modernc.org/sqlite"

	"chatHeaven/src"
)

var db *sql.DB

func main() {
	dbPath := "chatHeaven.db"

	var err error
	db, err = sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	_, err = db.ExecContext(context.Background(), `CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        username TEXT NOT NULL UNIQUE,
        email TEXT NOT NULL UNIQUE,
        mot_de_passe TEXT NOT NULL,
        profile_picture TEXT,
        user_likes INTEGER,
		createdAt TEXT,
		first_name TEXT,
		last_name TEXT
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
		date TEXT,
		FOREIGN KEY (user) REFERENCES users(username)
	)`)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.ExecContext(context.Background(), `CREATE TABLE IF NOT EXISTS posts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL UNIQUE,
		content TEXT NOT NULL,
		picture TEXT,
		user TEXT NOT NULL,
		topic TEXT NOT NULL,
		date TEXT,
		FOREIGN KEY (user) REFERENCES users(username)
		FOREIGN KEY (topic) REFERENCES topics(title)
	)`)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.ExecContext(context.Background(), `CREATE TABLE IF NOT EXISTS likes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user TEXT NOT NULL,
		title TEXT NOT NULL,
		FOREIGN KEY (user) REFERENCES users(username),
		FOREIGN KEY (title) REFERENCES posts(title)
	)`)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.ExecContext(context.Background(), `CREATE TABLE IF NOT EXISTS dislikes (
    	id INTEGER PRIMARY KEY AUTOINCREMENT,
    	user TEXT NOT NULL,
    	title TEXT NOT NULL,
    	FOREIGN KEY (user) REFERENCES users(username),
    	FOREIGN KEY (title) REFERENCES posts(title)
    )`)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.ExecContext(context.Background(), `CREATE TABLE IF NOT EXISTS comments (
    	id INTEGER PRIMARY KEY AUTOINCREMENT,
    	content TEXT NOT NULL,
    	user TEXT NOT NULL,
    	post TEXT NOT NULL,
		date TEXT,
    	FOREIGN KEY (user) REFERENCES users(username),
    	FOREIGN KEY (post) REFERENCES posts(title)
    )`)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", chatHeaven.HomeHandler)
	http.HandleFunc("/register", chatHeaven.RegisterHandler)
	http.HandleFunc("/login", chatHeaven.LoginHandler)
	http.HandleFunc("/user", chatHeaven.UserHandler)
	http.HandleFunc("/logout", chatHeaven.LogoutHandler)
	http.HandleFunc("/create-post", chatHeaven.CreatePost)
	http.HandleFunc("/posts", chatHeaven.PostsHandler)
	http.HandleFunc("/topics", chatHeaven.TopicsHandler)
	http.HandleFunc("/create-topic", chatHeaven.CreateTopic)
	http.HandleFunc("/post-content", chatHeaven.GetPostContent)
	http.HandleFunc("/like-post", chatHeaven.AddLike)
	http.HandleFunc("/dislike-post", chatHeaven.AddDislike)
	http.HandleFunc("/update-user", chatHeaven.UpdateUser)
	http.HandleFunc("/delete-post", chatHeaven.DeletePost)
	http.HandleFunc("/delete-topic", chatHeaven.DeleteTopic)
	http.HandleFunc("/search_autocomplete", chatHeaven.SearchAutocomplete)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	log.Println("Server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
