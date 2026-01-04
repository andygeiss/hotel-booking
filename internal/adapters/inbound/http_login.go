package inbound

import (
	"net/http"
	"os"

	"github.com/andygeiss/cloud-native-utils/templating"
)

// HttpViewLoginResponse specifies the view data.
type HttpViewLoginResponse struct {
	AppName string
	Title   string
}

// HttpViewLogin defines an HTTP handler function for rendering the login template.
func HttpViewLogin(e *templating.Engine) http.HandlerFunc {
	// Retrieve application details from environment variables at startup.
	// We can reuse these values instead of reading them from the environment on each request.
	appName := os.Getenv("APP_NAME")
	title := appName + " - " + os.Getenv("APP_DESCRIPTION")

	// Create the Data Object (DTO) once at startup.
	data := HttpViewLoginResponse{
		AppName: appName,
		Title:   title,
	}

	return func(w http.ResponseWriter, r *http.Request) {
		HttpView(e, "login", data)(w, r)
	}
}
