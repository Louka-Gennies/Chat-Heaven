package chatHeaven

type Topic struct {
	Title       string
	Description string
	NbPosts     int
	User        string
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
	Picture               string
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

type User struct {
	Username       string
	Email          string
	FirstName      string
	LastName       string
	ProfilePicture string
	LastPost       *LastPost
	TotalLikes     int
}
