package inbound

import "net/http"

// HttpViewErrorResponse specifies the view data for error pages.
type HttpViewErrorResponse struct {
	Layout
	ErrorTitle   string
	ErrorMessage string
	ErrorDetails string
}

// HttpViewError defines an HTTP handler function for rendering the error page.
// It reads error information from query parameters: title, message, and details.
func HttpViewError(v *View, app AppInfo) http.HandlerFunc {
	appName := app.Name
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

		v.Render(w, r, http.StatusOK, "error", "", data)
	}
}
