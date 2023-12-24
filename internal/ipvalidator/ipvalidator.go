package ipvalidator

import (
	"fmt"
	"net"
	"net/http"
	"strings"
)

type ValidatorIP struct {
	trustedSubnet *net.IPNet
}

func Create(trustedSubnet *net.IPNet) *ValidatorIP {
	return &ValidatorIP{
		trustedSubnet: trustedSubnet,
	}
}

func (v *ValidatorIP) ValidateIPHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if v.trustedSubnet == nil {
			next.ServeHTTP(w, r)
			return
		}

		ip, err := resolveIP(r)
		if err != nil {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		if !v.trustedSubnet.Contains(ip) {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func resolveIP(r *http.Request) (net.IP, error) {
	ipStr := r.Header.Get("X-Real-Ip")
	if ipStr == "" {
		return nil, fmt.Errorf("missing X-Real-IP header in request")
	}

	ip := net.ParseIP(ipStr)
	if ip == nil {
		ips := r.Header.Get("X-Forwarded-For")
		ipStrs := strings.Split(ips, ",")
		if len(ipStrs) != 0 {
			ipStr = ipStrs[0]
			ip = net.ParseIP(ipStr)
		}
	}

	if ip == nil {
		return nil, fmt.Errorf("failed parse ip from http header")
	}

	return ip, nil
}
