package main

import (
	"github.com/gorilla/mux"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestConsentRoutesRequireAuthentication(t *testing.T) {
	router := mux.NewRouter()
	registerAdminRoutes(router)
	for _, path := range []string{"/api/admin/oauth/consent"} {
		for _, method := range []string{"GET", "PUT"} {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(method, path, strings.NewReader(`{}`))
			r.Header.Set("Accept", "application/json")
			router.ServeHTTP(w, r)
			if w.Code != 401 && w.Code != 403 && w.Code != 302 {
				t.Fatal(path, method, w.Code)
			}
		}
	}
}
