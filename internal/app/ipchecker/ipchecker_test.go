// Package ipchecker for x-real-ip check.
package ipchecker_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/valinurovdenis/urlshortener/internal/app/ipchecker"
)

func TestIPChecker_CheckFromInnerNetwork(t *testing.T) {
	testCases := []struct {
		name    string
		ip      string
		network string
		status  int
	}{
		{name: "forbidden", ip: "192.168.2.1", network: "192.168.1.0/24", status: http.StatusForbidden},
		{name: "empty", ip: "", network: "192.168.1.0/24", status: http.StatusForbidden},
		{name: "ok", ip: "192.168.1.2", network: "192.168.1.0/24", status: http.StatusOK},
		{name: "wrong network", ip: "192.168.1.0/24", network: "kulebaka.168", status: http.StatusForbidden},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ipchecker := ipchecker.NewIPChecker(tc.network)
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
			handlerToTest := ipchecker.CheckFromInnerNetwork(nextHandler)

			req := httptest.NewRequest("GET", "http://testing", nil)
			req.Header.Set("X-Real-IP", tc.ip)
			resp := httptest.NewRecorder()
			defer resp.Result().Body.Close()
			handlerToTest.ServeHTTP(resp, req)
			require.Equal(t, tc.status, resp.Code)
		})
	}
}
