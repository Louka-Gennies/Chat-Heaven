package handler

import (
	"context"
	"database/sql"
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

        rows, err = db.QueryContext(context.Background(), "SELECT title, content, user, topic FROM posts WHERE topic = ? LIMIT ?", topicTitle, nbOfPosts[0])
    } else {
        rows, err = db.QueryContext(context.Background(), "SELECT title, content, user, topic FROM posts WHERE topic = ?", topicTitle)
    }

    if err != nil {
        return nil
    }
    defer rows.Close()

    var posts []Post
    for rows.Next() {
        var post Post
        if err := rows.Scan(&post.Title, &post.Content, &post.User, &post.Topic); err != nil {
            return nil
        }
        posts = append(posts, post)
    }
    return posts
}