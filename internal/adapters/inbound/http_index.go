package inbound

import (
	"net/http"

	"github.com/andygeiss/cloud-native-utils/redirecting"
	"github.com/andygeiss/cloud-native-utils/security"
	"github.com/andygeiss/cloud-native-utils/templating"
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
func HttpViewIndex(e *templating.Engine) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Make a shortcut for the current context.
		ctx := r.Context()

		// Check if the user is authenticated.
		if ctx.Value(security.ContextSessionID) == nil ||
			ctx.Value(security.ContextSessionID).(string) == "" {
			redirecting.Redirect(w, r, "/ui/login")
			return
		}

		// Add session-specific data.
		data := HttpViewIndexResponse{
			AppName:   "go-server",
			Email:     ctx.Value(security.ContextEmail).(string),
			Issuer:    ctx.Value(security.ContextIssuer).(string),
			Name:      ctx.Value(security.ContextName).(string),
			SessionID: ctx.Value(security.ContextSessionID).(string),
			Subject:   ctx.Value(security.ContextSubject).(string),
			Title:     "Home",
			Verified:  ctx.Value(security.ContextVerified).(bool),
		}

		// Render the template using the provided engine and data.
		HttpView(e, "index", data)(w, r)
	}
}
