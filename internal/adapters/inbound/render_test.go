package inbound_test

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/hotel-booking/internal/adapters/inbound"
	"github.com/andygeiss/hotel-booking/internal/domain/shared"
)

// ============================================================================
// The fragment-or-full-page test
// ============================================================================

// renderReservations runs the list handler with whatever htmx headers the case
// needs, and hands back the body.
func renderReservations(t *testing.T, headers map[string]string) (*httptest.ResponseRecorder, string) {
	t.Helper()

	handler := inbound.HttpViewReservations(newTestView(t), testApp(), createTestReservationService(t))
	req := httptest.NewRequest(http.MethodGet, "/ui/reservations", nil)
	for name, value := range headers {
		req.Header.Set(name, value)
	}
	req = addAuthContext(req, "test-session-123", "test@example.com")
	rec := httptest.NewRecorder()

	handler(rec, req)

	body, err := io.ReadAll(rec.Body)
	if err != nil {
		t.Fatalf("failed to read body: %v", err)
	}
	return rec, string(body)
}

func Test_Render_Without_Htmx_Headers_Should_Return_The_Whole_Page(t *testing.T) {
	// Arrange, Act
	rec, body := renderReservations(t, nil)

	// Assert
	assert.That(t, "status code must be 200", rec.Code, http.StatusOK)
	assert.That(t, "the document must be there", strings.Contains(body, "<!doctype html>"), true)
	assert.That(t, "the shell must be there", strings.Contains(body, "<main>"), true)
	assert.That(t, "the fragment must be inside it", strings.Contains(body, `id="reservation-list"`), true)
}

func Test_Render_With_Htmx_Request_Should_Return_Only_The_Fragment(t *testing.T) {
	// Arrange, Act
	rec, body := renderReservations(t, map[string]string{"HX-Request": "true"})

	// Assert
	assert.That(t, "status code must be 200", rec.Code, http.StatusOK)
	assert.That(t, "the fragment must be there", strings.Contains(body, `id="reservation-list"`), true)
	assert.That(t, "no document may come back", strings.Contains(body, "<!doctype html>"), false)
	assert.That(t, "no shell may come back", strings.Contains(body, "<main>"), false)
}

func Test_Render_With_Boosted_Request_Should_Return_The_Whole_Page(t *testing.T) {
	// A boosted link sends HX-Request: true but swaps the whole body, so it
	// needs the document. Answer it with a bare fragment and hx-boost
	// navigation renders an empty page — which is why isFragment tests both
	// headers rather than just the first.

	// Arrange, Act
	rec, body := renderReservations(t, map[string]string{"HX-Request": "true", "HX-Boosted": "true"})

	// Assert
	assert.That(t, "status code must be 200", rec.Code, http.StatusOK)
	assert.That(t, "the document must come back", strings.Contains(body, "<!doctype html>"), true)
	assert.That(t, "the shell must come back", strings.Contains(body, "<main>"), true)
}

// ============================================================================
// Caching
// ============================================================================

func Test_Render_Should_Vary_On_Both_Htmx_Headers(t *testing.T) {
	// Both headers pick the body, so both must be cache keys. Varying on
	// HX-Request alone lets a cache hand a bare fragment to a boosted
	// navigation.

	// Arrange, Act
	rec, _ := renderReservations(t, nil)

	// Assert
	vary := rec.Header().Values("Vary")
	joined := strings.Join(vary, ", ")
	assert.That(t, "Vary must name HX-Request", strings.Contains(joined, "HX-Request"), true)
	assert.That(t, "Vary must name HX-Boosted", strings.Contains(joined, "HX-Boosted"), true)
}

func Test_Render_Should_Add_To_Vary_Rather_Than_Replace_It(t *testing.T) {
	// Set would drop a Vary an earlier layer already added — Vary: Cookie from
	// the session middleware — and a cache could then serve one signed-in
	// user's page to another.

	// Arrange
	handler := inbound.HttpViewReservations(newTestView(t), testApp(), createTestReservationService(t))
	req := httptest.NewRequest(http.MethodGet, "/ui/reservations", nil)
	req = addAuthContext(req, "test-session-123", "test@example.com")
	rec := httptest.NewRecorder()
	rec.Header().Add("Vary", "Cookie")

	// Act
	handler(rec, req)

	// Assert
	joined := strings.Join(rec.Header().Values("Vary"), ", ")
	assert.That(t, "the earlier Vary must survive", strings.Contains(joined, "Cookie"), true)
	assert.That(t, "the new one must be there too", strings.Contains(joined, "HX-Request"), true)
}

// ============================================================================
// Escaping
// ============================================================================

func Test_Render_Should_Escape_Guest_Data(t *testing.T) {
	// These pages render a guest's own name, email and phone. html/template
	// escapes them for the context they land in; text/template would hand every
	// one of them to the browser as markup.

	// Arrange
	v := newTestView(t)
	data := inbound.HttpViewErrorResponse{
		AppName: "TestApp", Title: "TestApp - Error",
		ErrorTitle:   "Not Found",
		ErrorMessage: `<script>alert("xss")</script>`,
		ErrorDetails: `" onmouseover="alert(1)`,
	}
	req := httptest.NewRequest(http.MethodGet, "/ui/error", nil)
	rec := httptest.NewRecorder()

	// Act
	v.Render(rec, req, http.StatusOK, "error", "", data)

	// Assert
	body, _ := io.ReadAll(rec.Body)
	bodyStr := string(body)
	assert.That(t, "the script tag must not survive as markup", strings.Contains(bodyStr, "<script>alert"), false)
	assert.That(t, "it must come back escaped", strings.Contains(bodyStr, "&lt;script&gt;"), true)
	assert.That(t, "the attribute break-out must not survive", strings.Contains(bodyStr, `" onmouseover="`), false)
}

// ============================================================================
// Boot
// ============================================================================

func Test_NewView_Without_Any_Page_Should_Fail(t *testing.T) {
	// A glob that matches nothing means the templates moved and every page is
	// about to 500. Fail the process at boot instead.

	// Arrange, Act
	_, err := inbound.NewView(fstest.MapFS{}, slog.New(slog.DiscardHandler))

	// Assert
	assert.That(t, "error must not be nil", err != nil, true)
}

// ============================================================================
// Mutation flows: fragment, 303, HX-Redirect
// ============================================================================

// cancelReservation posts a cancellation with whatever htmx headers the case
// needs, against one seeded, still-cancellable reservation.
func cancelReservation(t *testing.T, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()

	repo := newMockReservationRepository()
	service := createDetailTestService(repo)
	checkIn := time.Now().AddDate(0, 0, 7).Truncate(24 * time.Hour)
	checkOut := checkIn.AddDate(0, 0, 3)
	repo.reservations[shared.ReservationID("res-001")] = *createTestReservation("res-001", "test@example.com", "room-101", checkIn, checkOut)

	handler := inbound.HttpCancelReservation(newTestView(t), testApp(), service)
	req := httptest.NewRequest(http.MethodPost, "/ui/reservations/res-001/cancel", nil)
	req.SetPathValue("id", "res-001")
	for name, value := range headers {
		req.Header.Set(name, value)
	}
	req = addAuthContext(req, "test-session-123", "test@example.com")
	rec := httptest.NewRecorder()

	handler(rec, req)
	return rec
}

func Test_Cancel_Without_Htmx_Should_Redirect(t *testing.T) {
	// POST-redirect-GET. A direct 200 would leave this POST's URL in history,
	// and the next refresh or Back would re-issue it as a GET against a
	// POST-only route.

	// Arrange, Act
	rec := cancelReservation(t, nil)

	// Assert
	assert.That(t, "status code must be 303", rec.Code, http.StatusSeeOther)
	assert.That(t, "it must point back at the list", rec.Header().Get("Location"), "/ui/reservations")
}

func Test_Cancel_When_Boosted_Should_Redirect_Too(t *testing.T) {
	// A boosted form is still a form: it needs the 303 for the same reason.

	// Arrange, Act
	rec := cancelReservation(t, map[string]string{"HX-Request": "true", "HX-Boosted": "true"})

	// Assert
	assert.That(t, "status code must be 303", rec.Code, http.StatusSeeOther)
}

func Test_Cancel_From_The_List_Should_Return_The_Fragment(t *testing.T) {
	// The list names a target, so the swap has somewhere to land.

	// Arrange, Act
	rec := cancelReservation(t, map[string]string{"HX-Request": "true", "HX-Target": "reservation-list"})

	// Assert
	assert.That(t, "status code must be 200", rec.Code, http.StatusOK)
	body, _ := io.ReadAll(rec.Body)
	assert.That(t, "the list fragment must come back", strings.Contains(string(body), `id="reservation-list"`), true)
	assert.That(t, "no document may come back", strings.Contains(string(body), "<!doctype html>"), false)
}

func Test_Cancel_From_The_Detail_Page_Should_Send_HX_Redirect(t *testing.T) {
	// The detail page has no target to swap into, and its whole document is
	// stale once the booking is gone, so send the browser somewhere else.

	// Arrange, Act
	rec := cancelReservation(t, map[string]string{"HX-Request": "true"})

	// Assert
	assert.That(t, "status code must be 200", rec.Code, http.StatusOK)
	assert.That(t, "it must redirect the client", rec.Header().Get("HX-Redirect"), "/ui/reservations")
}

// postReservation submits the create form with whatever htmx headers the case
// needs. The body is deliberately incomplete, so validation always rejects it.
func postInvalidReservation(t *testing.T, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()

	handler := inbound.HttpCreateReservation(newTestView(t), testApp(), createTestReservationService(t))
	req := httptest.NewRequest(http.MethodPost, "/ui/reservations", strings.NewReader("room_id=room-101"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for name, value := range headers {
		req.Header.Set(name, value)
	}
	req = addAuthContext(req, "test-session-123", "test@example.com")
	rec := httptest.NewRecorder()

	handler(rec, req)
	return rec
}

func Test_Invalid_Form_Should_Return_422_With_The_Form_Fragment(t *testing.T) {
	// Arrange, Act
	rec := postInvalidReservation(t, map[string]string{"HX-Request": "true"})

	// Assert
	assert.That(t, "status code must be 422", rec.Code, http.StatusUnprocessableEntity)
	body, _ := io.ReadAll(rec.Body)
	assert.That(t, "the form fragment must come back", strings.Contains(string(body), `id="reservation-form"`), true)
	assert.That(t, "no document may come back", strings.Contains(string(body), "<!doctype html>"), false)
}

func Test_Invalid_Boosted_Form_Should_Refuse_To_Push_The_Post_Url(t *testing.T) {
	// A boosted swap pushes the request URL into history by default. Let it push
	// this POST's URL and the next refresh or Back re-issues it as a GET against
	// a POST-only route — a 405. A 422 is not a redirect, so nothing else stops
	// the push; the header does.

	// Arrange, Act
	rec := postInvalidReservation(t, map[string]string{"HX-Request": "true", "HX-Boosted": "true"})

	// Assert
	assert.That(t, "status code must be 422", rec.Code, http.StatusUnprocessableEntity)
	assert.That(t, "the push must be refused", rec.Header().Get("HX-Push-Url"), "false")
}

func Test_Invalid_Plain_Form_Should_Not_Send_The_Push_Header(t *testing.T) {
	// Nothing is pushing anything without htmx, so the header would be noise.

	// Arrange, Act
	rec := postInvalidReservation(t, nil)

	// Assert
	assert.That(t, "status code must be 422", rec.Code, http.StatusUnprocessableEntity)
	assert.That(t, "no push header", rec.Header().Get("HX-Push-Url"), "")
}
