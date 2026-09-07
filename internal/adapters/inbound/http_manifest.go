package inbound

import (
	"encoding/json"
	"net/http"
)

// HttpViewManifestResponse specifies the view data for the PWA manifest.
//
// The manifest is JSON, so it is marshalled rather than templated: html/template
// escapes for an HTML context and would turn an ampersand in the application
// name into an entity inside a JSON string.
type HttpViewManifestResponse struct {
	BackgroundColor string         `json:"background_color"`
	Description     string         `json:"description"`
	Display         string         `json:"display"`
	Icons           []ManifestIcon `json:"icons"`
	Name            string         `json:"name"`
	Orientation     string         `json:"orientation"`
	Scope           string         `json:"scope"`
	ShortName       string         `json:"short_name"`
	StartURL        string         `json:"start_url"`
	ThemeColor      string         `json:"theme_color"`
}

// ManifestIcon is one entry of the manifest's icons array.
type ManifestIcon struct {
	Purpose string `json:"purpose"`
	Sizes   string `json:"sizes"`
	Src     string `json:"src"`
	Type    string `json:"type"`
}

// HttpViewManifest defines an HTTP handler function for serving the PWA manifest.json.
func HttpViewManifest(app AppInfo) http.HandlerFunc {
	// Marshal once at startup: the manifest depends on configuration only, and
	// nothing in it can change between requests.
	body, err := json.Marshal(HttpViewManifestResponse{
		BackgroundColor: "#000000",
		Description:     app.Description,
		Display:         "standalone",
		Icons: []ManifestIcon{
			{Purpose: "any", Sizes: "192x192", Src: "/static/img/icon-192.png", Type: "image/png"},
			{Purpose: "any", Sizes: "512x512", Src: "/static/img/icon-512.png", Type: "image/png"},
		},
		Name:        app.Name,
		Orientation: "portrait",
		Scope:       "/",
		ShortName:   app.Name,
		StartURL:    "/ui/",
		ThemeColor:  "#ffffff",
	})
	if err != nil {
		// Unreachable: every field is a string or a slice of strings.
		panic("inbound: marshalling the PWA manifest: " + err.Error())
	}

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/manifest+json")
		_, _ = w.Write(body)
	}
}
