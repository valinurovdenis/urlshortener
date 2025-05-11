// Package ipchecker for x-real-ip check.
package ipchecker

import (
	"log"
	"net"
	"net/http"
)

// Class to check inner network requests.
type IPChecker struct {
	netip net.IPNet
}

// Returns new checker.
// Requires CIDR network address.
func NewIPChecker(netaddress string) *IPChecker {
	_, netip, err := net.ParseCIDR(netaddress)
	if err != nil {
		log.Printf("cannot parse inner network ip")
		return &IPChecker{}
	}
	return &IPChecker{netip: *netip}
}

// Middleware checks if request from inner network.
func (a IPChecker) CheckFromInnerNetwork(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ipStr := r.Header.Get("X-Real-IP")
		ip := net.ParseIP(ipStr)
		if ip == nil || !a.netip.Contains(ip) {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		h.ServeHTTP(w, r)
	})
}
