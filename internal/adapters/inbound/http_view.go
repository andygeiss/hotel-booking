package inbound

import (
	"net/http"

	"github.com/andygeiss/cloud-native-utils/templating"
)

// HttpView defines an HTTP handler function for rendering a template with data.
// We use the templating engine from the cloud-native-utils package.
// We do not test the HttpView function because it's already tested in the cloud-native-utils package.
func HttpView(e *templating.Engine, name string, data any) http.HandlerFunc {
	return e.View(name, data)
}
