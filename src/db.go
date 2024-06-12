package chatHeaven

import (
	"database/sql"

	"github.com/gorilla/sessions"
	_ "modernc.org/sqlite"
)

var db *sql.DB
var store = sessions.NewCookieStore([]byte("menu-classique-burger"))