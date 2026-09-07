package inbound

import (
	"net/http"

	"github.com/andygeiss/cloud-native-utils/templating"
	"github.com/andygeiss/cloud-native-utils/web"
)

// HttpViewIndexResponse specifies the view data.
type HttpViewIndexResponse struct {
	AppName   string
	Email     string
	Issuer    string
	Name      string
	SessionID string
	Subject   string
	Title     string
	Verified  bool
}

// HttpViewIndex defines an HTTP handler function for rendering the index template.
func HttpViewIndex(e *templating.Engine, app AppInfo) http.HandlerFunc {
	appName := app.Name
	title := app.Title()

	return func(w http.ResponseWriter, r *http.Request) {
		// Make a shortcut for the current context.
		ctx := r.Context()

		// Check if the user is authenticated.
		// We check both sessionID and email because:
		// - sessionID might exist (from cookie) even after logout
		// - email being empty indicates the session was deleted server-side
		sessionID, _ := ctx.Value(web.ContextSessionID).(string)
		email, _ := ctx.Value(web.ContextEmail).(string)
		if sessionID == "" || email == "" {
			http.Redirect(w, r, "/ui/login", http.StatusSeeOther)
			return
		}

		// Add session-specific data.
		data := HttpViewIndexResponse{
			AppName:   appName,
			Email:     email,
			Issuer:    ctx.Value(web.ContextIssuer).(string),
			Name:      ctx.Value(web.ContextName).(string),
			SessionID: sessionID,
			Subject:   ctx.Value(web.ContextSubject).(string),
			Title:     title,
			Verified:  ctx.Value(web.ContextVerified).(bool),
		}

		// Render the template using the provided engine and data.
		HttpView(e, "index", data)(w, r)
	}
}
