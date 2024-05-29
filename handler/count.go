package handler

import (
	"context"
)


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
