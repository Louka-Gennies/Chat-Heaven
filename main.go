package main

import (
	"context"
	"database/sql"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)

var db *sql.DB
var store = sessions.NewCookieStore([]byte("menu-classique-burger"))

type Topic struct {
	Title       string
	Description string
	NbPosts     int
}

type Post struct {
	ID                    int
	Title                 string
	Content               string
	User                  string
	Topic                 string
	Likes                 int
	Dislikes              int
	NbComments            int
	LikeDislikeDifference int
	AlreadyLiked          bool
	AlreadyDisliked       bool
}

type Comment struct {
	Content   string
	User      string
	PostTitle string
}

func main() {
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
		title TEXT NOT NULL,
		description TEXT NOT NULL,
		picture TEXT,
		user TEXT NOT NULL,
		topic_likes INTEGER,
		FOREIGN KEY (user) REFERENCES users(username)
	)`)

	_, err = db.ExecContext(context.Background(), `CREATE TABLE IF NOT EXISTS likes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user TEXT NOT NULL,
		title TEXT NOT NULL,
		FOREIGN KEY (user) REFERENCES users(username),
		FOREIGN KEY (title) REFERENCES posts(title)
	)`)

	_, err = db.ExecContext(context.Background(), `CREATE TABLE IF NOT EXISTS dislikes (
    		id INTEGER PRIMARY KEY AUTOINCREMENT,
    		user TEXT NOT NULL,
    		title TEXT NOT NULL,
    		FOREIGN KEY (user) REFERENCES users(username),
    		FOREIGN KEY (title) REFERENCES posts(title)
    	)`)

	_, err = db.ExecContext(context.Background(), `CREATE TABLE IF NOT EXISTS posts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		content TEXT NOT NULL,
		picture TEXT,
		user TEXT NOT NULL,
		topic TEXT NOT NULL,
		FOREIGN KEY (user) REFERENCES users(username)
		FOREIGN KEY (topic) REFERENCES topics(title)
	)`)

	_, err = db.ExecContext(context.Background(), `CREATE TABLE IF NOT EXISTS comments (
    		id INTEGER PRIMARY KEY AUTOINCREMENT,
    		content TEXT NOT NULL,
    		user TEXT NOT NULL,
    		post TEXT NOT NULL,
    		FOREIGN KEY (user) REFERENCES users(username),
    		FOREIGN KEY (post) REFERENCES posts(title)
    	)`)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/user", userHandler)
	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/create-post", createPost)
	http.HandleFunc("/posts", postsHandler)
	http.HandleFunc("/topics", topicsHandler)
	http.HandleFunc("/create-topic", createTopic)
	http.HandleFunc("/post-content", getPostContent)
	http.HandleFunc("/like-post", addLike)
	http.HandleFunc("/dislike-post", addDislike)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	log.Println("Server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	username, ok := session.Values["username"]

	if !ok {
		tmpl, err := template.ParseFiles("templates/home.html")
		if err != nil {
			http.Error(w, "Error reading the HTML file", http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, nil)
		return
	}

	var profilePicture string
	err := db.QueryRowContext(context.Background(), "SELECT profile_picture FROM users WHERE username = ?", username).Scan(&profilePicture)
	if err != nil {
		http.Error(w, "Error retrieving the profile picture", http.StatusInternalServerError)
		return
	}

	data := struct {
		Username       string
		ProfilePicture string
		NbPosts        int
		NbTopics       int
		NbUsers        int
		Last4Topics    []Topic
	}{
		Username:       username.(string),
		ProfilePicture: profilePicture,
		NbPosts:        countPosts(),
		NbTopics:       countTopics(),
		NbUsers:        countUsers(),
		Last4Topics:    getTopics(4),
	}

	tmpl, err := template.ParseFiles("templates/home.html")
	if err != nil {
		http.Error(w, "Error reading the HTML file", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, data)
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		email := r.FormValue("email")
		password := r.FormValue("password")

		err := addUser(username, email, password, "./static/uploads/blank-pfp.png")
		if err != nil {
			http.Error(w, "Error during registration", http.StatusInternalServerError)
			return
		}

		session, _ := store.Get(r, "session")
		session.Values["username"] = username
		session.Save(r, w)

		http.Redirect(w, r, fmt.Sprintf("/user?username=%s", username), http.StatusSeeOther)

		return
	}

	tmpl, err := template.ParseFiles("templates/register.html")
	if err != nil {
		http.Error(w, "Error reading the HTML file", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		http.ServeFile(w, r, "templates/login.html")
	} else if r.Method == "POST" {
		username := r.FormValue("username")
		password := r.FormValue("password")

		err := verifyUser(username, password)
		if err != nil {
			http.Error(w, "Incorrect username or password", http.StatusUnauthorized)
			return
		}

		session, _ := store.Get(r, "session")
		session.Values["username"] = username
		session.Save(r, w)

		http.Redirect(w, r, fmt.Sprintf("/user?username=%s", username), http.StatusSeeOther)
	}
}

func userHandler(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	if username == "" {
		http.Error(w, "User not specified", http.StatusBadRequest)
		return
	}

	if r.Method == "GET" {
		var email, profilePicture string
		query := `SELECT email, profile_picture FROM users WHERE username = ?`
		err := db.QueryRowContext(context.Background(), query, username).Scan(&email, &profilePicture)
		if err != nil {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}

		data := struct {
			Username       string
			Email          string
			ProfilePicture string
		}{
			Username:       username,
			Email:          email,
			ProfilePicture: profilePicture,
		}

		tmpl, err := template.ParseFiles("templates/user.html")
		if err != nil {
			http.Error(w, "Error reading the HTML file", http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, data)
	} else if r.Method == "POST" {
		file, handler, err := r.FormFile("profile_picture")
		if err != nil {
			http.Error(w, "Error during file upload", http.StatusInternalServerError)
			return
		}
		defer file.Close()

		os.MkdirAll("static/uploads", os.ModePerm)

		filePath := filepath.Join("static/uploads", handler.Filename)
		f, err := os.Create(filePath)
		if err != nil {
			http.Error(w, "Error saving the file", http.StatusInternalServerError)
			return
		}
		defer f.Close()
		io.Copy(f, file)

		updateSQL := `UPDATE users SET profile_picture = ? WHERE username = ?`
		_, err = db.ExecContext(context.Background(), updateSQL, "/static/uploads/"+handler.Filename, username)
		if err != nil {
			http.Error(w, "Error updating the profile picture", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/user?username=%s", username), http.StatusSeeOther)
	}
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	session.Options.MaxAge = -1
	session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func addUser(username, email, motDePasse, profilePicture string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(motDePasse), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = db.ExecContext(context.Background(), `INSERT INTO users (username, email, mot_de_passe, profile_picture) VALUES (?, ?, ?, ?)`,
		username, email, hashedPassword, profilePicture)
	if err != nil {
		return err
	}
	return nil
}

func verifyUser(username, motDePasse string) error {
	var motDePasseDB string
	err := db.QueryRowContext(context.Background(), "SELECT mot_de_passe FROM users WHERE username = ?", username).Scan(&motDePasseDB)
	if err != nil {
		return err
	}

	err = bcrypt.CompareHashAndPassword([]byte(motDePasseDB), []byte(motDePasse))
	if err != nil {
		return fmt.Errorf("incorrect password")
	}
	return nil
}

func postsHandler(w http.ResponseWriter, r *http.Request) {
	topic := r.URL.Query().Get("topic")
	Posts := getPosts(topic)
	session, _ := store.Get(r, "session")
	username, ok := session.Values["username"]
	if !ok {
		tmpl, err := template.ParseFiles("templates/home.html")
		if err != nil {
			http.Error(w, "Error reading the HTML file", http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, nil)
		return
	}

	var profilePicture string
	err := db.QueryRowContext(context.Background(), "SELECT profile_picture FROM users WHERE username = ?", username).Scan(&profilePicture)
	if err != nil {
		http.Error(w, "Error retrieving the profile picture", http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles("templates/post.html")
	if err != nil {
		http.Error(w, "Error reading the HTML file", http.StatusInternalServerError)
		return
	}

	data := struct {
		Topic          string
		Post           []Post
		Username       string
		ProfilePicture string
	}{
		Topic:          topic,
		Post:           Posts,
		Username:       username.(string),
		ProfilePicture: profilePicture,
	}

	tmpl.Execute(w, data)
}

func createPost(w http.ResponseWriter, r *http.Request) {

	session, _ := store.Get(r, "session")
	username, ok := session.Values["username"]
	if !ok {
		tmpl, err := template.ParseFiles("templates/home.html")
		if err != nil {
			http.Error(w, "Error reading the HTML file", http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, nil)
		return
	}

	var profilePicture string
	err := db.QueryRowContext(context.Background(), "SELECT profile_picture FROM users WHERE username = ?", username).Scan(&profilePicture)
	if err != nil {
		http.Error(w, "Error retrieving the profile picture", http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles("templates/create-post.html")
	if err != nil {
		http.Error(w, "Error reading the HTML file", http.StatusInternalServerError)
		return
	}

	topic := r.URL.Query().Get("topic")

	data := struct {
		Topic          string
		Username       string
		ProfilePicture string
	}{
		Topic:          topic,
		Username:       username.(string),
		ProfilePicture: profilePicture,
	}

	if r.Method == "POST" {
		session, _ := store.Get(r, "session")
		username, ok := session.Values["username"]
		title, content, topicTitle := r.FormValue("title"), r.FormValue("content"), r.FormValue("topic")

		if !ok {
			http.Error(w, "You must be logged in to post a message", http.StatusUnauthorized)
			return
		}
		_, err := db.ExecContext(context.Background(), "INSERT INTO posts (user, title, content, topic) VALUES (?, ?, ?, ?)", username, title, content, topicTitle)
		if err != nil {
			fmt.Println(err)
			http.Error(w, "Error posting the message", http.StatusInternalServerError)
			return
		}
		fmt.Println("Post added successfully!")
		http.Redirect(w, r, fmt.Sprintf("/posts?topic=%s", topic), http.StatusSeeOther)

	}

	tmpl.Execute(w, data)
}

func topicsHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	username, ok := session.Values["username"]
	if !ok {
		tmpl, err := template.ParseFiles("templates/home.html")
		if err != nil {
			http.Error(w, "Error reading the HTML file", http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, nil)
		return
	}

	var profilePicture string
	err := db.QueryRowContext(context.Background(), "SELECT profile_picture FROM users WHERE username = ?", username).Scan(&profilePicture)
	if err != nil {
		http.Error(w, "Error retrieving the profile picture", http.StatusInternalServerError)
		return
	}

	rows, err := db.QueryContext(context.Background(), "SELECT title, description FROM topics")
	if err != nil {
		http.Error(w, "Error retrieving the messages", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	tmpl, err := template.ParseFiles("templates/topic.html")
	if err != nil {
		http.Error(w, "Error reading the HTML file", http.StatusInternalServerError)
		return
	}

	var Topics []Topic
	for rows.Next() {
		var topic Topic
		if err := rows.Scan(&topic.Title, &topic.Description); err != nil {
			http.Error(w, "Error reading the messages", http.StatusInternalServerError)
			return
		}
		topic.NbPosts = len(getPosts(topic.Title))
		Topics = append(Topics, topic)
	}

	data := struct {
		Topic          []Topic
		Username       string
		ProfilePicture string
	}{
		Topic:          Topics,
		Username:       username.(string),
		ProfilePicture: profilePicture,
	}

	tmpl.Execute(w, data)
}

func createTopic(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	username, ok := session.Values["username"]
	if !ok {
		tmpl, err := template.ParseFiles("templates/home.html")
		if err != nil {
			http.Error(w, "Error reading the HTML file", http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, nil)
		return
	}

	var profilePicture string
	err := db.QueryRowContext(context.Background(), "SELECT profile_picture FROM users WHERE username = ?", username).Scan(&profilePicture)
	if err != nil {
		http.Error(w, "Error retrieving the profile picture", http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles("templates/create-topic.html")
	if err != nil {
		http.Error(w, "Error reading the HTML file", http.StatusInternalServerError)
		return
	}

	if r.Method == "POST" {

		session, _ := store.Get(r, "session")

		username, ok := session.Values["username"]

		title, description := r.FormValue("title"), r.FormValue("description")

		if !ok {
			http.Error(w, "You must be logged in to post a message", http.StatusUnauthorized)
			return
		}
		_, err := db.ExecContext(context.Background(), "INSERT INTO topics (user, title, description) VALUES (?, ?, ?)", username, title, description)
		if err != nil {
			fmt.Println(err)
			http.Error(w, "Error posting the message", http.StatusInternalServerError)
			return
		}

		fmt.Println("Topic added successfully!")
		http.Redirect(w, r, "/topics", http.StatusSeeOther)

	}
	data := struct {
		Username       string
		ProfilePicture string
	}{
		Username:       username.(string),
		ProfilePicture: profilePicture,
	}

	tmpl.Execute(w, data)
}

func countPosts() int {
	var count int
	err := db.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM posts").Scan(&count)
	if err != nil {
		return 0
	}
	return count
}

func countTopics() int {
	var count int
	err := db.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM topics").Scan(&count)
	if err != nil {
		return 0
	}
	return count
}

func countUsers() int {
	var count int
	err := db.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		return 0
	}
	return count
}

func getTopics(nbOfTopics int) []Topic {
	totalTopics := countTopics()
	if nbOfTopics > totalTopics {
		nbOfTopics = totalTopics
	}
	if nbOfTopics > 0 {
		rows, err := db.QueryContext(context.Background(), "SELECT title, Description FROM topics")
		if err != nil {
			return nil
		}
		defer rows.Close()

		topics := make([]Topic, 0, nbOfTopics)
		for rows.Next() {
			var topic Topic
			if err := rows.Scan(&topic.Title, &topic.Description); err != nil {
				return nil
			}
			topics = append(topics, topic)
		}
		return topics
	} else {
		rows, err := db.QueryContext(context.Background(), "SELECT title, description FROM topics")
		if err != nil {
			return nil
		}
		defer rows.Close()

		var topics []Topic
		for rows.Next() {
			var topic Topic
			if err := rows.Scan(&topic.Title, &topic.Description); err != nil {
				return nil
			}

			topics = append(topics, topic)
		}
		return topics
	}
}

func getPosts(topicTitle string, nbOfPosts ...int) []Post {
	var rows *sql.Rows
	var err error

	if len(nbOfPosts) > 0 && nbOfPosts[0] > 0 {
		totalPosts := countPosts()
		if nbOfPosts[0] > totalPosts {
			nbOfPosts[0] = totalPosts
		}

		rows, err = db.QueryContext(context.Background(), "SELECT id,title, content, user, topic FROM posts WHERE topic = ? LIMIT ?", topicTitle, nbOfPosts[0])
	} else {
		rows, err = db.QueryContext(context.Background(), "SELECT id,title, content, user, topic FROM posts WHERE topic = ?", topicTitle)
	}

	if err != nil {
		return nil
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		if err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.User, &post.Topic); err != nil {
			return nil
		}
		post.Likes = likeCount(post.ID)
		post.Dislikes = dislikeCount(post.ID)
		post.NbComments = len(getComment(post.Title))
		posts = append(posts, post)
	}
	return posts
}

func getComment(title string) []Comment {
	rows, err := db.QueryContext(context.Background(), "SELECT content, user FROM comments WHERE post = ?", title)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		var comment Comment
		if err := rows.Scan(&comment.Content, &comment.User); err != nil {
			return nil
		}
		comment.PostTitle = title
		comments = append(comments, comment)
	}
	return comments
}

func likeCount(postID int) int {
	var count int
	err := db.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM likes WHERE title = ?", postID).Scan(&count)
	if err != nil {
		return 0
	}
	return count

}

func dislikeCount(postID int) int {
	var count int
	err := db.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM dislikes WHERE title = ?", postID).Scan(&count)
	if err != nil {
		return 0
	}
	return count

}

func getPostContent(w http.ResponseWriter, r *http.Request) {
	postID := r.URL.Query().Get("postID")
	postIDInt, err := strconv.Atoi(postID)
	session, _ := store.Get(r, "session")
	username, ok := session.Values["username"]
	if !ok {
		tmpl, err := template.ParseFiles("templates/home.html")
		if err != nil {
			http.Error(w, "Error reading the HTML file", http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, nil)
		return
	}

	var profilePicture string
	err = db.QueryRowContext(context.Background(), "SELECT profile_picture FROM users WHERE username = ?", username).Scan(&profilePicture)
	if err != nil {
		http.Error(w, "Error retrieving the profile picture", http.StatusInternalServerError)
		return
	}
	if err != nil {
		// handle error
		fmt.Println(err)
	}
	if postID == "" {
		http.Error(w, "Message not specified", http.StatusBadRequest)
		return
	}

	var title, content, user, topic string
	err = db.QueryRowContext(context.Background(), "SELECT title, content, user, topic FROM posts WHERE id = ?", postIDInt).Scan(&title, &content, &user, &topic)
	if err != nil {
		http.Error(w, "Message not found", http.StatusNotFound)
		return
	}

	if r.Method == "POST" {
		session, _ := store.Get(r, "session")
		username, ok := session.Values["username"]
		comment := r.FormValue("content")
		if !ok {
			http.Error(w, "You must be logged in to comment on a message", http.StatusUnauthorized)
			return
		}
		_, err := db.ExecContext(context.Background(), "INSERT INTO comments (user, content, post) VALUES (?, ?, ?)", username, comment, title)
		if err != nil {
			http.Error(w, "Error posting the comment", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, fmt.Sprintf("/post-content?postID=%d", postIDInt), http.StatusSeeOther)

	}

	data := struct {
		Post           Post
		Comment        []Comment
		Username       string
		ProfilePicture string
	}{
		Post: Post{
			ID:                    postIDInt,
			Title:                 title,
			Content:               content,
			User:                  user,
			Topic:                 topic,
			Likes:                 likeCount(postIDInt),
			Dislikes:              dislikeCount(postIDInt),
			LikeDislikeDifference: likeCount(postIDInt) - dislikeCount(postIDInt),
			AlreadyLiked:          isLiked(username.(string), postIDInt),
			AlreadyDisliked:       isDisliked(username.(string), postIDInt),
		},
		Comment:        getComment(title),
		Username:       username.(string),
		ProfilePicture: profilePicture,
	}

	tmpl, err := template.ParseFiles("templates/post-content.html")
	if err != nil {
		http.Error(w, "Error reading the HTML file", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, data)
}

func addLike(w http.ResponseWriter, r *http.Request) {
	postID := r.URL.Query().Get("postID")
	postIDInt, err := strconv.Atoi(postID)
	if err != nil {
		// handle error
		fmt.Println(err)
	}
	if postID == "" {
		http.Error(w, "Message not specified", http.StatusBadRequest)
		return
	}

	session, _ := store.Get(r, "session")
	username, ok := session.Values["username"]
	if !ok {
		http.Error(w, "You must be logged in to like a message", http.StatusUnauthorized)
		return
	}

	var existingLike int
	var existingDislike int
	err = db.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM likes WHERE user = ? AND title = ?", username, postIDInt).Scan(&existingLike)
	err = db.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM dislikes WHERE user = ? AND title = ?", username, postIDInt).Scan(&existingDislike)
	if err != nil {
		http.Error(w, "Error checking the likes", http.StatusInternalServerError)
		return
	}

	if existingLike > 0 {
		// If a like exists, delete it
		_, err = db.ExecContext(context.Background(), "DELETE FROM likes WHERE user = ? AND title = ?", username, postIDInt)
		if err != nil {
			http.Error(w, "Error deleting the like", http.StatusInternalServerError)
			return
		}
	} else if existingDislike > 0 {
		// If no like exists, add a new one
		_, err = db.ExecContext(context.Background(), "INSERT INTO likes (user, title) VALUES (?, ?)", username, postIDInt)
		_, err = db.ExecContext(context.Background(), "DELETE FROM dislikes WHERE user = ? AND title = ?", username, postIDInt)
		if err != nil {
			http.Error(w, "Error adding the like", http.StatusInternalServerError)
			return
		}
	} else {
		_, err = db.ExecContext(context.Background(), "INSERT INTO likes (user, title) VALUES (?, ?)", username, postIDInt)
		if err != nil {
			http.Error(w, "Error adding the like", http.StatusInternalServerError)
			return
		}
	}

	http.Redirect(w, r, fmt.Sprintf("/post-content?postID=%d", postIDInt), http.StatusSeeOther)
}

func addDislike(w http.ResponseWriter, r *http.Request) {
	postID := r.URL.Query().Get("postID")
	postIDInt, err := strconv.Atoi(postID)
	if err != nil {
		// handle error
		fmt.Println(err)
	}
	if postID == "" {
		http.Error(w, "Message not specified", http.StatusBadRequest)
		return
	}

	session, _ := store.Get(r, "session")
	username, ok := session.Values["username"]
	if !ok {
		http.Error(w, "You must be logged in to like a message", http.StatusUnauthorized)
		return
	}

	var existingDislike int
	var existingLike int
	err = db.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM dislikes WHERE user = ? AND title = ?", username, postIDInt).Scan(&existingDislike)
	err = db.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM likes WHERE user = ? AND title = ?", username, postIDInt).Scan(&existingLike)
	if err != nil {
		http.Error(w, "Error checking the dislikes", http.StatusInternalServerError)
		return
	}

	if existingDislike > 0 {
		// If a dislike exists, delete it
		_, err = db.ExecContext(context.Background(), "DELETE FROM dislikes WHERE user = ? AND title = ?", username, postIDInt)
		if err != nil {
			http.Error(w, "Error deleting the dislike", http.StatusInternalServerError)
			return
		}
	} else if existingLike > 0 {
		// If no dislike exists, add a new one
		_, err = db.ExecContext(context.Background(), "INSERT INTO dislikes (user, title) VALUES (?, ?)", username, postIDInt)
		_, err = db.ExecContext(context.Background(), "DELETE FROM likes WHERE user = ? AND title = ?", username, postIDInt)
		if err != nil {
			http.Error(w, "Error adding the dislike", http.StatusInternalServerError)
			return
		}
	} else {
		_, err = db.ExecContext(context.Background(), "INSERT INTO dislikes (user, title) VALUES (?, ?)", username, postIDInt)
		if err != nil {
			http.Error(w, "Error adding the dislike", http.StatusInternalServerError)
			return
		}
	}

	http.Redirect(w, r, fmt.Sprintf("/post-content?postID=%d", postIDInt), http.StatusSeeOther)
}

func isLiked(username string, postID int) bool {
	var existingLike int
	err := db.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM likes WHERE user = ? AND title = ?", username, postID).Scan(&existingLike)
	if err != nil {
		return false
	}
	return existingLike > 0
}

func isDisliked(username string, postID int) bool {
	var existingDislike int
	err := db.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM dislikes WHERE user = ? AND title = ?", username, postID).Scan(&existingDislike)
	if err != nil {
		return false
	}
	return existingDislike > 0
}
