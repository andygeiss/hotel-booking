package inbound

import (
	"net/http"

	"github.com/andygeiss/cloud-native-utils/templating"
)

// HttpViewLoginResponse specifies the view data.
type HttpViewLoginResponse struct {
	AppName string
	Title   string
}

// HttpViewLogin defines an HTTP handler function for rendering the login template.
func HttpViewLogin(e *templating.Engine) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := HttpViewLoginResponse{
			AppName: "go-server",
			Title:   "Login",
		}
		HttpView(e, "login", data)(w, r)
	}
}
