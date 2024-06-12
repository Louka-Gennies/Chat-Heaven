package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

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
	User 	    string
	LastPost    *LastPost
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
	Date                  string
}

type Comment struct {
	CommentID int
	Content   string
	User      string
	PostTitle string
	Date      string
}

type LastPost struct {
	Title  string
	Author string
	Date   string
	ID     int
}

type Result struct {
	Type  string `json:"type"`
	Value string `json:"value"`
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
	http.HandleFunc("/update-user", updateUser)
	http.HandleFunc("/delete-post", deletePost)
	http.HandleFunc("/delete-topic", deleteTopic)
	http.HandleFunc("/search_autocomplete", searchAutocomplete)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	log.Println("Server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "session")
	if err != nil {
    http.Error(w, "Error retrieving session", http.StatusInternalServerError)
    return
    }

	username, ok := session.Values["username"]

	if !ok {
		data := struct {
			Username       string
			ProfilePicture string
			NbPosts        int
			NbTopics       int
			NbUsers        int
			Last4Topics    []Topic
			IsLogged       bool
		}{
			Username:       "",
			ProfilePicture: "",
			NbPosts:        countPosts(),
			NbTopics:       countTopics(),
			NbUsers:        countUsers(),
			Last4Topics:    getTopics(3),
			IsLogged:       false,
		}
	
		tmpl, err := template.ParseFiles("templates/home.html")
		if err != nil {
			http.Error(w, "Error reading the HTML file", http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, data)
		return
	} else if ok {
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
			IsLogged       bool
		}{
			Username:       username.(string),
			ProfilePicture: profilePicture,
			NbPosts:        countPosts(),
			NbTopics:       countTopics(),
			NbUsers:        countUsers(),
			Last4Topics:    getTopics(3),
			IsLogged:       true,
		}

		tmpl, err := template.ParseFiles("templates/home.html")
		if err != nil {
			http.Error(w, "Error reading the HTML file", http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, data)
	}
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		email := r.FormValue("email")
		password := r.FormValue("password")
		date := time.Now().Format("02-01-2006")

		err := addUser(username, email, password, "./static/uploads/blank-pfp.png", date)
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
		var email, profilePicture, createdAt, first_name, last_name string
		query := `SELECT email, profile_picture, first_name, last_name, createdAt FROM users WHERE username = ?`
		err := db.QueryRowContext(context.Background(), query, username).Scan(&email, &profilePicture, &first_name, &last_name, &createdAt)
		if err != nil {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}

		data := struct {
			Username       string
			Email          string
			ProfilePicture string
			FirstName      string
			LastName       string
			CreatedAt 	   string
			Posts		  []Post
			Topics		  []Topic
			CreatedAt      string
		}{
			Username:       username,
			Email:          email,
			ProfilePicture: profilePicture,
			FirstName:      first_name,
			LastName:       last_name,
			CreatedAt: 	    createdAt,
			Posts:          getPostsByUser(username),
			Topics:         getTopicByUser(username),
			CreatedAt:      createdAt,
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

func updateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		username := r.URL.Query().Get("username")
		firstName := r.FormValue("first_name")
		lastName := r.FormValue("last_name")

		updateSQL := `UPDATE users SET first_name = ?, last_name = ? WHERE username = ?`
		_, err := db.ExecContext(context.Background(), updateSQL, firstName, lastName, username)
		if err != nil {
			http.Error(w, "Error updating the user", http.StatusInternalServerError)
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

func addUser(username, email, motDePasse, profilePicture, date string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(motDePasse), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = db.ExecContext(context.Background(), `INSERT INTO users (username, email, mot_de_passe, profile_picture, first_name, last_name, createdAt) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		username, email, hashedPassword, profilePicture, "", "", date)
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
		Topic            string
		TopicDescription string
		Post             []Post
		Username         string
		ProfilePicture   string
	}{
		Topic:            topic,
		TopicDescription: getTopicDescription(topic),
		Post:             Posts,
		Username:         username.(string),
		ProfilePicture:   profilePicture,
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
		date := time.Now().Format("02-01-2006 15:04")

		if !ok {
			http.Error(w, "You must be logged in to post a message", http.StatusUnauthorized)
			return
		}
		_, err := db.ExecContext(context.Background(), "INSERT INTO posts (user, title, content, topic, date) VALUES (?, ?, ?, ?, ?)", username, title, content, topicTitle, date)
		if err != nil {
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
		topic.LastPost = getLastPost(topic.Title)
		Topics = append(Topics, topic)
	}
	fmt.Println(Topics)

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

		date := time.Now().Format("02-01-2006 15:04")

		if !ok {
			http.Error(w, "You must be logged in to post a message", http.StatusUnauthorized)
			return
		}
		_, err := db.ExecContext(context.Background(), "INSERT INTO topics (user, title, description, date) VALUES (?, ?, ?, ?)", username, title, description, date)
		if err != nil {
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
		post.Date = getDatePost(post.ID)
		posts = append(posts, post)
	}
	return posts
}

func getTopicByUser(username string) []Topic {
	rows, err := db.QueryContext(context.Background(), "SELECT title, description, user FROM topics WHERE user = ?", username)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var topics []Topic
	for rows.Next() {
		var topic Topic
		if err := rows.Scan(&topic.Title, &topic.Description, &topic.User); err != nil {
			return nil
		}
		topics = append(topics, topic)
	}
	return topics
}

func getPostsByUser(username string) []Post {
	rows, err := db.QueryContext(context.Background(), "SELECT id, title, content, user, topic FROM posts WHERE user = ?", username)
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
		if err := rows.Scan(&comment.CommentID, &comment.Content, &comment.User); err != nil {
			return nil
		}
		comment.Date = getDateComment(comment.CommentID)
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
	if err != nil {
		http.Error(w, "Error converting the post ID", http.StatusInternalServerError)
		return
	}
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
		date := time.Now().Format("02-01-2006 15:04")
		if !ok {
			http.Error(w, "You must be logged in to comment on a message", http.StatusUnauthorized)
			return
		}
		_, err := db.ExecContext(context.Background(), "INSERT INTO comments (user, content, post, date) VALUES (?, ?, ?, ?)", username, comment, title, date)
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
			Date:                  getDatePost(postIDInt),
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
	if err != nil {
		http.Error(w, "Error checking the likes", http.StatusInternalServerError)
		return
	}
	err = db.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM dislikes WHERE user = ? AND title = ?", username, postIDInt).Scan(&existingDislike)
	if err != nil {
		http.Error(w, "Error checking the dislikes", http.StatusInternalServerError)
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
		if err != nil {
			http.Error(w, "Error adding the like", http.StatusInternalServerError)
			return
		}
		_, err = db.ExecContext(context.Background(), "DELETE FROM dislikes WHERE user = ? AND title = ?", username, postIDInt)
		if err != nil {
			http.Error(w, "Error adding the dislike", http.StatusInternalServerError)
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
	if err != nil {
		http.Error(w, "Error checking the dislikes", http.StatusInternalServerError)
		return
	}
	err = db.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM likes WHERE user = ? AND title = ?", username, postIDInt).Scan(&existingLike)
	if err != nil {
		http.Error(w, "Error checking the likes", http.StatusInternalServerError)
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
		if err != nil {
			http.Error(w, "Error adding the dislike", http.StatusInternalServerError)
			return
		}
		_, err = db.ExecContext(context.Background(), "DELETE FROM likes WHERE user = ? AND title = ?", username, postIDInt)
		if err != nil {
			http.Error(w, "Error adding the like", http.StatusInternalServerError)
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

func deletePost(w http.ResponseWriter, r *http.Request) {
	postTitle := r.URL.Query().Get("post")
	if postTitle == "" {
		http.Error(w, "Post not specified", http.StatusBadRequest)
		return
	}

	_, err := db.ExecContext(context.Background(), "DELETE FROM posts WHERE title = ?", postTitle)
	if err != nil {
		http.Error(w, "Error deleting the post", http.StatusInternalServerError)
		return
	}

	username := r.URL.Query().Get("user")


	http.Redirect(w, r, fmt.Sprintf("/user?username=%s", username), http.StatusSeeOther)
}

func deleteTopic(w http.ResponseWriter, r *http.Request) {
	topicTitle := r.URL.Query().Get("topic")
	if topicTitle == "" {
		http.Error(w, "Topic not specified", http.StatusBadRequest)
		return
	}

	_, err := db.ExecContext(context.Background(), "DELETE FROM topics WHERE title = ?", topicTitle)
	if err != nil {
		http.Error(w, "Error deleting the topic", http.StatusInternalServerError)
		return
	}

	username := r.URL.Query().Get("user")

	http.Redirect(w, r, fmt.Sprintf("/user?username=%s", username), http.StatusSeeOther)
}
func getTopicDescription(topic string) string {
	var description string
	err := db.QueryRowContext(context.Background(), "SELECT description FROM topics WHERE title = ?", topic).Scan(&description)
	if err != nil {
		return ""
	}
	return description
}

func getDatePost(postID int) string {
	var date string
	err := db.QueryRowContext(context.Background(), "SELECT date FROM posts WHERE id = ?", postID).Scan(&date)
	if err != nil {
		return ""
	}
	return date
}

func getDateComment(commentID int) string {
	var date string
	err := db.QueryRowContext(context.Background(), "SELECT date FROM comments WHERE id = ?", commentID).Scan(&date)
	if err != nil {
		return ""
	}
	return date
}

func getLastPost(topic string) *LastPost {
	var title string
	err := db.QueryRowContext(context.Background(), "SELECT title FROM posts WHERE topic = ? ORDER BY id DESC LIMIT 1", topic).Scan(&title)
	if err != nil {
		return nil
	}

	var lastPost LastPost
	err = db.QueryRowContext(context.Background(), "SELECT title, user, date, id FROM posts WHERE title = ?", title).Scan(&lastPost.Title, &lastPost.Author, &lastPost.Date, &lastPost.ID)
	if err != nil {
		return nil
	}

	return &lastPost
}

func searchAutocomplete(w http.ResponseWriter, r *http.Request) {
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
