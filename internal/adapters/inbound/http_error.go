package inbound

import (
	"net/http"
	"os"

	"github.com/andygeiss/cloud-native-utils/templating"
)

// HttpViewErrorResponse specifies the view data for error pages.
type HttpViewErrorResponse struct {
	AppName      string
	Title        string
	ErrorTitle   string
	ErrorMessage string
	ErrorDetails string
}

// HttpViewError defines an HTTP handler function for rendering the error template.
// It reads error information from query parameters: title, message, and details.
func HttpViewError(e *templating.Engine) http.HandlerFunc {
	// Retrieve application details from environment variables at startup.
	appName := os.Getenv("APP_NAME")
	pageTitle := appName + " - Error"

	return func(w http.ResponseWriter, r *http.Request) {
		// Read error details from query parameters.
		errorTitle := r.URL.Query().Get("title")
		errorMessage := r.URL.Query().Get("message")
		errorDetails := r.URL.Query().Get("details")

		// Set defaults if not provided.
		if errorTitle == "" {
			errorTitle = "An Error Occurred"
		}
		if errorMessage == "" {
			errorMessage = "Something went wrong. Please try again."
		}

		data := HttpViewErrorResponse{
			AppName:      appName,
			Title:        pageTitle,
			ErrorTitle:   errorTitle,
			ErrorMessage: errorMessage,
			ErrorDetails: errorDetails,
		}

		HttpView(e, "error", data)(w, r)
	}
}
