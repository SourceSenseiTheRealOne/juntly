package httpapi_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/httpapi"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/listingmedia"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/users"
	"github.com/google/uuid"
)

func TestMediaRoutesProtectScopeAndNeverExposeStorageReferences(t *testing.T) {
	reader := &httpMediaReader{}
	finalizer := &httpMediaFinalizer{}
	handler := httpapi.NewMediaRouter(http.NotFoundHandler(), staticVerifier{identity: users.VerifiedIdentity{Subject: "owner"}}, reader, finalizer)
	listing := uuid.NewString()
	media := uuid.NewString()
	for _, test := range []struct {
		path, method, body string
		token              bool
		want               int
	}{
		{"/api/v1/me/listings/" + listing + "/media", "GET", "", false, 401},
		{"/api/v1/moderation/listings/" + listing + "/media", "GET", "", false, 401},
		{"/api/v1/me/listings/" + listing + "/media/" + media + "/finalize", "POST", "{}", false, 401},
		{"/api/v1/public/listings/" + listing + "/media", "GET", "", false, 200},
		{"/api/v1/me/listings/" + listing + "/media", "GET", "", true, 200},
		{"/api/v1/moderation/listings/" + listing + "/media", "GET", "", true, 200},
		{"/api/v1/me/listings/" + listing + "/media/" + media + "/finalize", "POST", "{}", true, 204},
		{"/api/v1/me/listings/" + listing + "/media/" + media + "/finalize", "POST", `{"reference":"evil"}`, true, 400},
		{"/api/v1/public/listings/not-a-uuid/media", "GET", "", false, 400},
		{"/api/v1/public/listings/" + listing + "/media?scope=moderator", "GET", "", false, 400},
	} {
		r := httptest.NewRequest(test.method, test.path, strings.NewReader(test.body))
		if test.token {
			r.Header.Set("Authorization", "Bearer synthetic-token")
		}
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		if w.Code != test.want || strings.Contains(w.Body.String(), "objectReference") {
			t.Fatalf("%s returned %d instead of %d", test.path, w.Code, test.want)
		}
	}
	if finalizer.calls != 1 || reader.calls != 3 || reader.scope != listingmedia.ModeratorRead {
		t.Fatal("invalid/authless request reached service or route scope was not authoritative")
	}
}

func TestMediaRoutesServeNoStorePNGAndBoundErrors(t *testing.T) {
	reader := &httpMediaReader{body: []byte("synthetic-image-bytes")}
	finalizer := &httpMediaFinalizer{err: listingmedia.ErrConflict}
	handler := httpapi.NewMediaRouter(http.NotFoundHandler(), staticVerifier{identity: users.VerifiedIdentity{Subject: "owner"}}, reader, finalizer)
	path := "/api/v1/public/listings/" + uuid.NewString() + "/media/" + uuid.NewString()
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
	if w.Code != 200 || w.Header().Get("Content-Type") != "image/png" || w.Header().Get("Cache-Control") != "private, no-store" || w.Header().Get("X-Content-Type-Options") != "nosniff" || w.Body.String() != string(reader.body) {
		t.Fatal("unsafe image HTTP response")
	}
	for err, status := range map[error]int{listingmedia.ErrNotFound: 404, errors.New("private provider key"): 503} {
		reader.err = err
		w = httptest.NewRecorder()
		handler.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		if w.Code != status || strings.Contains(w.Body.String(), "private provider key") {
			t.Fatal("unsafe error response")
		}
	}
	r := httptest.NewRequest("POST", strings.Replace(path, "/public/", "/me/", 1)+"/finalize", nil)
	r.Header.Set("Authorization", "Bearer synthetic-token")
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	if w.Code != 409 {
		t.Fatal("revision conflict lost")
	}
}

type httpMediaReader struct {
	calls int
	scope listingmedia.ReadScope
	body  []byte
	err   error
}

func (r *httpMediaReader) List(_ context.Context, _ users.VerifiedIdentity, scope listingmedia.ReadScope, _ uuid.UUID) ([]listingmedia.Photo, error) {
	r.calls++
	r.scope = scope
	return []listingmedia.Photo{}, r.err
}
func (r *httpMediaReader) Get(context.Context, users.VerifiedIdentity, listingmedia.ReadScope, uuid.UUID, uuid.UUID) ([]byte, error) {
	r.calls++
	return r.body, r.err
}

type httpMediaFinalizer struct {
	calls int
	err   error
}

func (f *httpMediaFinalizer) Finalize(context.Context, users.VerifiedIdentity, uuid.UUID, uuid.UUID) error {
	f.calls++
	return f.err
}
