// Package ipchecker for x-real-ip check.
package ipchecker

import (
	"context"
	"log"
	"net"
	"net/http"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
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

// Interceptor checks if request from inner network.
func (a IPChecker) GrpcCheckFromInner(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (interface{}, error) {

	md, ok := metadata.FromIncomingContext(ctx)
	realIP := md.Get("X-Real-IP")
	if ok && len(realIP) > 0 {
		if ip := net.ParseIP(realIP[0]); ip != nil && a.netip.Contains(ip) {
			return handler(ctx, req)
		}
	}
	return nil, status.Errorf(codes.PermissionDenied, "Handler only for inner use")
}
