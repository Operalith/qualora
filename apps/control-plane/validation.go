package main

import (
	"context"
	"fmt"
	"net"
	"net/netip"
	"net/url"
	"strings"
	"time"
)

type targetResolver interface {
	LookupIPAddr(ctx context.Context, host string) ([]net.IPAddr, error)
}

func NormalizeProjectRequest(input CreateProjectRequest) (CreateProjectRequest, error) {
	return NormalizeProjectRequestWithResolver(input, net.DefaultResolver)
}

func NormalizeProjectRequestWithResolver(input CreateProjectRequest, resolver targetResolver) (CreateProjectRequest, error) {
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

	input.FrontendURL = strings.TrimSpace(input.FrontendURL)
	input.APIBaseURL = strings.TrimSpace(input.APIBaseURL)
	input.OpenAPIURL = strings.TrimSpace(input.OpenAPIURL)
	if input.FrontendURL == "" && input.APIBaseURL == "" && input.OpenAPIURL == "" {
		return input, fmt.Errorf("at least one of frontend_url, api_base_url, or openapi_url is required")
	}

	allowedHosts, err := NormalizeAllowedHosts(input.AllowedHosts, input.AllowPrivateTargets)
	if err != nil {
		return input, err
	}
	input.AllowedHosts = allowedHosts

	if input.FrontendURL != "" {
		frontendURL, err := validateTargetURL(input.FrontendURL, allowedHosts, input.AllowPrivateTargets, resolver)
		if err != nil {
			return input, fmt.Errorf("frontend_url: %w", err)
		}
		input.FrontendURL = frontendURL.String()
	}

	if input.APIBaseURL != "" {
		apiURL, err := validateTargetURL(input.APIBaseURL, allowedHosts, input.AllowPrivateTargets, resolver)
		if err != nil {
			return input, fmt.Errorf("api_base_url: %w", err)
		}
		input.APIBaseURL = apiURL.String()
	}

	if input.OpenAPIURL != "" {
		openAPIURL, err := validateTargetURL(input.OpenAPIURL, allowedHosts, input.AllowPrivateTargets, resolver)
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
		if (parsed.Path != "" && parsed.Path != "/") || parsed.RawQuery != "" || parsed.Fragment != "" {
			return "", fmt.Errorf("allowed host %q must not include a path, query, or fragment", input)
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
	return validateTargetURL(raw, allowedHosts, allowPrivateTargets, net.DefaultResolver)
}

func validateTargetURL(raw string, allowedHosts []string, allowPrivateTargets bool, resolver targetResolver) (*url.URL, error) {
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
	if err := validateResolvedTarget(host, allowPrivateTargets, resolver); err != nil {
		return nil, err
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
	if !allowPrivateTargets && isMetadataHost(host) {
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

func validateResolvedTarget(host string, allowPrivateTargets bool, resolver targetResolver) error {
	if allowPrivateTargets || resolver == nil || net.ParseIP(host) != nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	records, err := resolver.LookupIPAddr(ctx, host)
	if err != nil {
		return fmt.Errorf("host %q could not be resolved by DNS", host)
	}
	if len(records) == 0 {
		return fmt.Errorf("host %q did not resolve to any IP addresses", host)
	}

	for _, record := range records {
		addr, ok := netip.AddrFromSlice(record.IP)
		if !ok {
			return fmt.Errorf("host %q resolved to an invalid IP address", host)
		}
		if isBlockedResolvedAddr(addr) {
			return fmt.Errorf("host %q resolves to a blocked private, loopback, link-local, multicast, unspecified, or metadata IP address", host)
		}
	}
	return nil
}

func isBlockedResolvedAddr(addr netip.Addr) bool {
	if addr.Is4In6() {
		addr = addr.Unmap()
	}
	return addr.IsPrivate() ||
		addr.IsLoopback() ||
		addr.IsLinkLocalUnicast() ||
		addr.IsLinkLocalMulticast() ||
		addr.IsMulticast() ||
		addr.IsUnspecified() ||
		addr.String() == "169.254.169.254" ||
		addr.String() == "100.100.100.200"
}

func isMetadataHost(host string) bool {
	switch host {
	case "metadata", "metadata.google.internal", "metadata.goog", "instance-data", "169.254.169.254", "100.100.100.200":
		return true
	default:
		return false
	}
}
