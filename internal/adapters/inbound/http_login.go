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
func HttpViewLogin(e *templating.Engine, app AppInfo) http.HandlerFunc {
	appName := app.Name
	title := app.Title()

	// Create the Data Object (DTO) once at startup.
	data := HttpViewLoginResponse{
		AppName: appName,
		Title:   title,
	}

	return func(w http.ResponseWriter, r *http.Request) {
		HttpView(e, "login", data)(w, r)
	}
}
