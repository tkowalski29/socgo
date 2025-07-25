package internal

import "net/http"

func GetUserID(r *http.Request) string {
	return "default_user"
}
