package inbound

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"path"
	"strings"
)

const (
	// layoutFile is the shell every page is rendered into. Its base name is
	// also the template name, because ParseFS names a file's content after it.
	layoutFile = "assets/templates/layout.tmpl"
	pagesGlob  = "assets/templates/pages/*.tmpl"
)

// Layout is what the shell needs from every page. Each page's view data embeds
// it, so layout.tmpl reads the same fields whichever page it wraps, and a page
// cannot forget to supply one.
type Layout struct {
	AppName   string
	Title     string
	SessionID string // empty means nobody is signed in: the nav offers Sign In
	ShowNew   bool   // the mobile action bar offers the New Reservation shortcut
}

// View renders every HTML page.
//
// It holds one parsed template set per page — the layout plus that page's file
// — rather than one set for all of them. Every page defines "title" and "main",
// so a single set would keep only the last page parsed and silently render the
// wrong body.
//
// html/template, never text/template: a guest's name, email and phone reach
// these templates. text/template escapes nothing, which hands each of them to
// the browser as markup.
type View struct {
	logger *slog.Logger
	pages  map[string]*template.Template
}

// NewView parses every page at boot, so a broken template fails the process
// rather than the first request that needs it.
func NewView(efs fs.FS, logger *slog.Logger) (*View, error) {
	names, err := fs.Glob(efs, pagesGlob)
	if err != nil {
		return nil, fmt.Errorf("globbing %q: %w", pagesGlob, err)
	}
	if len(names) == 0 {
		return nil, fmt.Errorf("no page templates matched %q", pagesGlob)
	}

	pages := make(map[string]*template.Template, len(names))
	for _, name := range names {
		t, err := template.New(path.Base(layoutFile)).ParseFS(efs, layoutFile, name)
		if err != nil {
			return nil, fmt.Errorf("parsing %q with the layout: %w", name, err)
		}
		pages[strings.TrimSuffix(path.Base(name), ".tmpl")] = t
	}
	return &View{logger: logger, pages: pages}, nil
}

// isFragment reports whether this request wants a fragment back rather than a
// whole page.
//
// Boosted links and forms send HX-Request: true but swap the entire body, so
// they need the full page — without the HX-Boosted half, hx-boost navigation
// renders bare fragments into empty pages. Every handler that chooses between a
// fragment and a 303 calls this too, so the rule has one definition and a
// handler can never disagree with its own rendering.
func isFragment(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true" && r.Header.Get("HX-Boosted") != "true"
}

// Render writes page as a whole document, or only its named block when the
// request came from an htmx interaction. An empty block always renders the full
// page.
func (v *View) Render(w http.ResponseWriter, r *http.Request, status int, page, block string, data any) {
	t, ok := v.pages[page]
	if !ok {
		v.fail(w, r, fmt.Errorf("no template set for page %q", page))
		return
	}

	name := path.Base(layoutFile)
	if block != "" && isFragment(r) {
		name = block
	}

	// Buffer first: a template that fails halfway after WriteHeader has already
	// sent a 200 and half a page, and there is no taking either back.
	var buf bytes.Buffer
	if err := t.ExecuteTemplate(&buf, name, data); err != nil {
		v.fail(w, r, fmt.Errorf("rendering %q from page %q: %w", name, page, err))
		return
	}

	// Both headers help pick the body, so both must be cache keys: Vary on
	// HX-Request alone lets a cache serve a bare fragment to a boosted
	// navigation. Add, never Set — Set would drop a Vary the session layer has
	// already added, letting a cache serve one user's page to another.
	w.Header().Add("Vary", "HX-Request, HX-Boosted")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = buf.WriteTo(w)
}

// fail logs what actually broke and tells the client nothing about it.
func (v *View) fail(w http.ResponseWriter, r *http.Request, err error) {
	if v.logger != nil {
		v.logger.Error("render failed", "error", err, "path", r.URL.Path)
	}
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}
