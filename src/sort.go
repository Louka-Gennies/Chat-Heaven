package chatHeaven

import (
	"time"
)

func sortByLikes(post []Post) []Post {
	for i := 0; i < len(post); i++ {
		for j := i + 1; j < len(post); j++ {
			if post[i].LikeDislikeDifference < post[j].LikeDislikeDifference {
				post[i], post[j] = post[j], post[i]
			}
		}
	}
	return post
}

func sortByDate(post []Post) []Post {
	const layout = "2006-01-02"
	for i := 0; i < len(post); i++ {
		for j := i + 1; j < len(post); j++ {
			iDate, _ := time.Parse(layout, post[i].Date)
			jDate, _ := time.Parse(layout, post[j].Date)
			if iDate.Before(jDate) {
				post[i], post[j] = post[j], post[i]
			}
		}
	}
	return post
}

func sortByComments(post []Post) []Post {
	for i := 0; i < len(post); i++ {
		for j := i + 1; j < len(post); j++ {
			if post[i].NbComments < post[j].NbComments {
				post[i], post[j] = post[j], post[i]
			}
		}
	}
	return post
}

func sortByAuthor(post []Post) []Post {
	for i := 0; i < len(post); i++ {
		for j := i + 1; j < len(post); j++ {
			if post[i].User > post[j].User {
				post[i], post[j] = post[j], post[i]
			}
		}
	}
	return post
}
