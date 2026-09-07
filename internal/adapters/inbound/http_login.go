package inbound

import "net/http"

// HttpViewLoginResponse specifies the view data.
type HttpViewLoginResponse struct {
	Layout
}

// HttpViewLogin defines an HTTP handler function for rendering the login page.
func HttpViewLogin(v *View, app AppInfo) http.HandlerFunc {
	// Create the Data Object (DTO) once at startup. SessionID stays empty, which
	// is what makes the layout render the signed-out navigation.
	data := HttpViewLoginResponse{
		AppName: app.Name,
		Title:   app.Title(),
	}

	return func(w http.ResponseWriter, r *http.Request) {
		v.Render(w, r, http.StatusOK, "login", "", data)
	}
}
