package inbound

// AppInfo is the application identity every page renders: the name in the
// header, the title in the browser tab, the version in the manifest.
//
// It is passed in rather than read from the environment. Handlers used to call
// os.Getenv("APP_NAME") themselves, which is why every handler test had to set
// that variable before it could render anything. The baseline puts every knob
// in one Config parsed in main, and keeps internal/ out of the environment
// entirely; this struct is how the two identity fields get here.
type AppInfo struct {
	Description string
	Name        string
	Version     string
}

// Title is what goes in the <title> element: the name, then the description.
func (a AppInfo) Title() string {
	return a.Name + " - " + a.Description
}
