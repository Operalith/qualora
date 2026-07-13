package main

import (
	"fmt"
	"net"
	"net/netip"
	"net/url"
	"strings"
)

func NormalizeProjectRequest(input CreateProjectRequest) (CreateProjectRequest, error) {
	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" {
		return input, fmt.Errorf("name is required")
	}

	if input.SecurityMode == "" {
		input.SecurityMode = "passive"
	}
	if input.SecurityMode != "passive" {
		return input, fmt.Errorf("security_mode must be passive for the MVP")
	}
	if input.DestructiveActions {
		return input, fmt.Errorf("destructive_actions must be false for the MVP")
	}

	allowedHosts, err := NormalizeAllowedHosts(input.AllowedHosts, input.AllowPrivateTargets)
	if err != nil {
		return input, err
	}
	input.AllowedHosts = allowedHosts

	frontendURL, err := ValidateTargetURL(input.FrontendURL, allowedHosts, input.AllowPrivateTargets)
	if err != nil {
		return input, fmt.Errorf("frontend_url: %w", err)
	}
	input.FrontendURL = frontendURL.String()

	if strings.TrimSpace(input.APIBaseURL) != "" {
		apiURL, err := ValidateTargetURL(input.APIBaseURL, allowedHosts, input.AllowPrivateTargets)
		if err != nil {
			return input, fmt.Errorf("api_base_url: %w", err)
		}
		input.APIBaseURL = apiURL.String()
	}

	if strings.TrimSpace(input.OpenAPIURL) != "" {
		openAPIURL, err := ValidateTargetURL(input.OpenAPIURL, allowedHosts, input.AllowPrivateTargets)
		if err != nil {
			return input, fmt.Errorf("openapi_url: %w", err)
		}
		input.OpenAPIURL = openAPIURL.String()
	}

	return input, nil
}

func NormalizeAllowedHosts(hosts []string, allowPrivateTargets bool) ([]string, error) {
	seen := make(map[string]struct{}, len(hosts))
	normalized := make([]string, 0, len(hosts))

	for _, host := range hosts {
		value, err := NormalizeAllowedHost(host)
		if err != nil {
			return nil, err
		}
		if isBlockedHost(value, allowPrivateTargets) {
			return nil, fmt.Errorf("allowed host %q is blocked by the default safety policy", value)
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		normalized = append(normalized, value)
	}

	if len(normalized) == 0 {
		return nil, fmt.Errorf("allowed_hosts must contain at least one host")
	}
	return normalized, nil
}

func NormalizeAllowedHost(input string) (string, error) {
	value := strings.ToLower(strings.TrimSpace(input))
	if value == "" {
		return "", fmt.Errorf("allowed_hosts contains an empty host")
	}

	if strings.Contains(value, "://") {
		parsed, err := url.Parse(value)
		if err != nil {
			return "", fmt.Errorf("allowed host %q is invalid", input)
		}
		value = parsed.Hostname()
	}

	if strings.ContainsAny(value, "/?#") {
		return "", fmt.Errorf("allowed host %q must not include a path, query, or fragment", input)
	}

	if host, _, err := net.SplitHostPort(value); err == nil {
		value = host
	}

	value = strings.Trim(value, "[]")
	value = strings.TrimSuffix(value, ".")
	if value == "" {
		return "", fmt.Errorf("allowed host %q is invalid", input)
	}
	return value, nil
}

func ValidateTargetURL(raw string, allowedHosts []string, allowPrivateTargets bool) (*url.URL, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return nil, fmt.Errorf("URL is required")
	}

	parsed, err := url.Parse(value)
	if err != nil {
		return nil, fmt.Errorf("URL is invalid")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, fmt.Errorf("only http and https URLs are supported")
	}

	host := strings.ToLower(strings.TrimSuffix(parsed.Hostname(), "."))
	if host == "" {
		return nil, fmt.Errorf("host is required")
	}
	if isBlockedHost(host, allowPrivateTargets) {
		return nil, fmt.Errorf("host %q is blocked by the default safety policy", host)
	}
	if !HostAllowed(host, allowedHosts) {
		return nil, fmt.Errorf("host %q is not present in allowed_hosts", host)
	}

	parsed.Scheme = strings.ToLower(parsed.Scheme)
	parsed.Host = strings.ToLower(parsed.Host)
	return parsed, nil
}

func HostAllowed(host string, allowedHosts []string) bool {
	host = strings.ToLower(strings.TrimSuffix(strings.TrimSpace(host), "."))
	for _, allowed := range allowedHosts {
		allowed = strings.ToLower(strings.TrimSuffix(strings.TrimSpace(allowed), "."))
		if host == allowed {
			return true
		}
		if strings.HasPrefix(allowed, "*.") && strings.HasSuffix(host, strings.TrimPrefix(allowed, "*")) {
			return true
		}
	}
	return false
}

func isBlockedHost(host string, allowPrivateTargets bool) bool {
	host = strings.ToLower(strings.TrimSuffix(strings.TrimSpace(host), "."))

	if !allowPrivateTargets {
		if host == "localhost" || strings.HasSuffix(host, ".localhost") || strings.HasSuffix(host, ".local") {
			return true
		}
	}

	// Common cloud metadata endpoints should never be reached by accident.
	if !allowPrivateTargets && (host == "metadata.google.internal" || host == "169.254.169.254" || host == "100.100.100.200") {
		return true
	}

	addr, err := netip.ParseAddr(host)
	if err != nil {
		return false
	}
	if allowPrivateTargets {
		return false
	}

	return addr.IsPrivate() ||
		addr.IsLoopback() ||
		addr.IsLinkLocalUnicast() ||
		addr.IsLinkLocalMulticast() ||
		addr.IsMulticast() ||
		addr.IsUnspecified()
}
