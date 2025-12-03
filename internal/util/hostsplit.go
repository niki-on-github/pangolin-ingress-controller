// Package util provides utility functions for the Pangolin Ingress Controller.
package util

import (
	"errors"
	"net"
	"strings"

	"golang.org/x/net/publicsuffix"
)

var (
	// ErrInvalidHost is returned when the host cannot be processed.
	ErrInvalidHost = errors.New("invalid host")

	// ErrWildcardHost is returned for wildcard hosts which are not supported.
	ErrWildcardHost = errors.New("wildcard hosts are not supported")

	// ErrIPAddress is returned when the host is an IP address.
	ErrIPAddress = errors.New("IP addresses are not supported as hosts")
)

// SplitHost splits a hostname into subdomain and domain components.
// Uses the public suffix list to correctly identify the registrable domain.
//
// Examples:
//   - "app.example.com" -> subdomain="app", domain="example.com"
//   - "api.staging.example.com" -> subdomain="api.staging", domain="example.com"
//   - "example.com" -> subdomain="", domain="example.com"
//   - "www.example.co.uk" -> subdomain="www", domain="example.co.uk"
func SplitHost(host string) (subdomain, domain string, err error) {
	// Validate input
	if host == "" {
		return "", "", ErrInvalidHost
	}

	// Check for wildcards
	if strings.Contains(host, "*") {
		return "", "", ErrWildcardHost
	}

	// Check for IP addresses
	if ip := net.ParseIP(host); ip != nil {
		return "", "", ErrIPAddress
	}

	// Check for localhost variants
	if host == "localhost" || strings.HasPrefix(host, "localhost:") {
		return "", "", ErrInvalidHost
	}

	// Get the effective TLD+1 (registrable domain)
	domain, err = publicsuffix.EffectiveTLDPlusOne(host)
	if err != nil {
		return "", "", ErrInvalidHost
	}

	// If host equals domain, there's no subdomain
	if host == domain {
		return "", domain, nil
	}

	// Extract subdomain (everything before the domain)
	subdomain = strings.TrimSuffix(host, "."+domain)

	return subdomain, domain, nil
}
