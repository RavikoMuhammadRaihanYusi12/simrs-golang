package delivery

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/thinkingx/simrs-backend/pkg/jwt"
)

// self-check for the guard: no token=401, wrong role=403, right role=200.
func TestGuard(t *testing.T) {
	jwtSvc := jwt.NewJWTService("test-secret")
	d := &Delivery{jwtService: jwtSvc}

	ok := func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }
	guarded := d.guard([]string{"doctor"}, ok)

	// no token
	rec := httptest.NewRecorder()
	guarded(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("no token: got %d want 401", rec.Code)
	}

	// nurse token -> forbidden
	nurseTok, _ := jwtSvc.Login("loket", "loket123")
	rec = httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+nurseTok)
	guarded(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("nurse: got %d want 403", rec.Code)
	}

	// doctor token -> ok
	docTok, _ := jwtSvc.Login("dokter", "dokter123")
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+docTok)
	guarded(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("doctor: got %d want 200", rec.Code)
	}
}
