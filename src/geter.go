package chatHeaven

import (
	"context"
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

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
			topic.LastPost = getLastPost(topic.Title)
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
			topic.LastPost = getLastPost(topic.Title)
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
	rows, err := db.QueryContext(context.Background(), "SELECT id, content, user FROM comments WHERE post = ?", title)
	fmt.Println(err)
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
