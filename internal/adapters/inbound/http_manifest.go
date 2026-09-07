package inbound

import (
	"net/http"

	"github.com/andygeiss/cloud-native-utils/templating"
)

// HttpViewManifestResponse specifies the view data for the PWA manifest.
type HttpViewManifestResponse struct {
	Description string
	Name        string
	ShortName   string
}

// HttpViewManifest defines an HTTP handler function for rendering the PWA manifest.json.
func HttpViewManifest(e *templating.Engine, app AppInfo) http.HandlerFunc {
	appName := app.Name
	description := app.Description

	// Create the Data Object (DTO) once at startup.
	data := HttpViewManifestResponse{
		Description: description,
		Name:        appName,
		ShortName:   appName,
	}

	return func(w http.ResponseWriter, r *http.Request) {
		// Set the content type to application/manifest+json for PWA manifest.
		w.Header().Set("Content-Type", "application/manifest+json")
		HttpView(e, "manifest", data)(w, r)
	}
}
