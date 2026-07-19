package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime"
	"net/url"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

const maxStoredSpecBytes = 2 * 1024 * 1024

var (
	openAPIHTTPMethods = map[string]struct{}{
		"GET": {}, "HEAD": {}, "OPTIONS": {}, "POST": {}, "PUT": {}, "PATCH": {}, "DELETE": {}, "TRACE": {},
	}
	openAPISafeMethods = map[string]struct{}{
		"GET": {}, "HEAD": {}, "OPTIONS": {},
	}
	openAPISensitivePathPattern = regexp.MustCompile(`(?i)(delete|remove|update|create|payment|transfer|admin|mutation|reset|token|password)`)
)

type parsedOpenAPISpec struct {
	Title      string
	Version    string
	ServerURL  string
	Operations []APIOperation
}

func ParseOpenAPISpec(raw string, sourceURL string, apiBaseURL string) (*parsedOpenAPISpec, error) {
	if len([]byte(raw)) > maxStoredSpecBytes {
		return nil, fmt.Errorf("OpenAPI document is too large for the current alpha import limit")
	}

	doc, err := decodeOpenAPIDocument(raw)
	if err != nil {
		return nil, err
	}
	version := stringField(doc, "openapi")
	if !strings.HasPrefix(version, "3.") {
		return nil, fmt.Errorf("only OpenAPI 3.x documents are supported in this alpha")
	}
	paths := mapField(doc, "paths")
	if len(paths) == 0 {
		return nil, fmt.Errorf("OpenAPI document must include a non-empty paths object")
	}

	info := mapField(doc, "info")
	rootSecurity, hasRootSecurity := doc["security"]
	servers := sliceField(doc, "servers")
	serverURL := resolveServerURL(stringField(firstMap(servers), "url"), sourceURL, apiBaseURL)

	operations := make([]APIOperation, 0)
	pathNames := make([]string, 0, len(paths))
	for pathName := range paths {
		pathNames = append(pathNames, pathName)
	}
	sort.Strings(pathNames)

	for _, pathName := range pathNames {
		pathItem := asMap(paths[pathName])
		if pathItem == nil {
			continue
		}
		methodNames := make([]string, 0, len(pathItem))
		for key := range pathItem {
			method := strings.ToUpper(key)
			if _, ok := openAPIHTTPMethods[method]; ok {
				methodNames = append(methodNames, method)
			}
		}
		sort.Strings(methodNames)
		for _, method := range methodNames {
			operation := asMap(pathItem[strings.ToLower(method)])
			if operation == nil {
				operation = asMap(pathItem[method])
			}
			if operation == nil {
				continue
			}
			operations = append(operations, classifyOperation(pathName, method, pathItem, operation, rootSecurity, hasRootSecurity, doc))
		}
	}

	return &parsedOpenAPISpec{
		Title:      stringField(info, "title"),
		Version:    stringField(info, "version"),
		ServerURL:  serverURL,
		Operations: operations,
	}, nil
}

func decodeOpenAPIDocument(raw string) (map[string]any, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, fmt.Errorf("OpenAPI document is empty")
	}

	var doc map[string]any
	decoder := json.NewDecoder(strings.NewReader(trimmed))
	decoder.UseNumber()
	if strings.HasPrefix(trimmed, "{") {
		if err := decoder.Decode(&doc); err != nil {
			return nil, fmt.Errorf("OpenAPI JSON could not be parsed: %w", err)
		}
		return normalizeMap(doc), nil
	}
	if err := yaml.NewDecoder(bytes.NewBufferString(trimmed)).Decode(&doc); err != nil {
		return nil, fmt.Errorf("OpenAPI YAML could not be parsed: %w", err)
	}
	return normalizeMap(doc), nil
}

func classifyOperation(pathName string, method string, pathItem map[string]any, operation map[string]any, rootSecurity any, hasRootSecurity bool, doc map[string]any) APIOperation {
	apiOperation := APIOperation{
		Method:               method,
		Path:                 pathName,
		OperationID:          stringField(operation, "operationId"),
		Summary:              stringField(operation, "summary"),
		Description:          stringField(operation, "description"),
		Tags:                 stringSlice(sliceField(operation, "tags")),
		ExpectedStatuses:     expectedStatuses(operation),
		ExpectedContentTypes: expectedContentTypes(operation),
		ResponseSchemas:      expectedResponseSchemas(operation, doc),
	}

	requiresAuth := operationRequiresAuthentication(operation, rootSecurity, hasRootSecurity)
	apiOperation.RequiresAuthentication = &requiresAuth

	if _, ok := openAPISafeMethods[method]; !ok {
		apiOperation.SkipReason = fmt.Sprintf("method %s is not read-only in this alpha", method)
		return apiOperation
	}
	if openAPISensitivePathPattern.MatchString(pathName) {
		apiOperation.SkipReason = "path appears sensitive or mutation-oriented"
		return apiOperation
	}
	if requestBodyRequired(operation) {
		apiOperation.SkipReason = "operation requires a request body"
		return apiOperation
	}

	resolvedPath, queryString, skipReason := resolveOperationTarget(pathName, pathItem, operation)
	if skipReason != "" {
		apiOperation.SkipReason = skipReason
		return apiOperation
	}
	apiOperation.ResolvedPath = resolvedPath
	apiOperation.QueryString = queryString
	if requiresAuth {
		apiOperation.SkipReason = "operation declares authentication requirements"
		return apiOperation
	}
	apiOperation.SafeToExecute = true
	return apiOperation
}

func operationRequiresAuthentication(operation map[string]any, rootSecurity any, hasRootSecurity bool) bool {
	if security, ok := operation["security"]; ok {
		return securityRequiresAuthentication(security)
	}
	if hasRootSecurity {
		return securityRequiresAuthentication(rootSecurity)
	}
	return false
}

func securityRequiresAuthentication(security any) bool {
	items := asSlice(security)
	if len(items) == 0 {
		return false
	}
	for _, item := range items {
		requirement := asMap(item)
		if len(requirement) == 0 {
			return false
		}
	}
	return true
}

func requestBodyRequired(operation map[string]any) bool {
	body := mapField(operation, "requestBody")
	return boolField(body, "required")
}

func resolveOperationTarget(pathName string, pathItem map[string]any, operation map[string]any) (string, string, string) {
	parameters := append(sliceField(pathItem, "parameters"), sliceField(operation, "parameters")...)
	resolvedPath := pathName

	for _, name := range pathParameterNames(pathName) {
		param := findParameter(parameters, "path", name)
		sample, ok := safeParameterSample(param)
		if !ok {
			return "", "", fmt.Sprintf("path parameter %s requires a safe example, default, or enum value", name)
		}
		resolvedPath = strings.ReplaceAll(resolvedPath, "{"+name+"}", url.PathEscape(sample))
	}

	query := url.Values{}
	for _, rawParam := range parameters {
		param := asMap(rawParam)
		if param == nil || stringField(param, "in") != "query" || !boolField(param, "required") {
			continue
		}
		name := stringField(param, "name")
		if name == "" {
			continue
		}
		if sensitiveAPIParameterName(name) {
			return "", "", fmt.Sprintf("required query parameter %s appears sensitive", name)
		}
		sample, ok := safeParameterSample(param)
		if !ok {
			return "", "", fmt.Sprintf("required query parameter %s requires a safe example, default, or enum value", name)
		}
		query.Set(name, sample)
	}
	return resolvedPath, query.Encode(), ""
}

func pathParameterNames(pathName string) []string {
	matches := regexp.MustCompile(`\{([^}]+)\}`).FindAllStringSubmatch(pathName, -1)
	names := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) == 2 && match[1] != "" {
			names = append(names, match[1])
		}
	}
	return names
}

func findParameter(parameters []any, location string, name string) map[string]any {
	var found map[string]any
	for _, rawParam := range parameters {
		param := asMap(rawParam)
		if param == nil {
			continue
		}
		if stringField(param, "in") == location && stringField(param, "name") == name {
			found = param
		}
	}
	return found
}

func safeParameterSample(param map[string]any) (string, bool) {
	if param == nil {
		return "", false
	}
	for _, value := range []any{param["example"], param["default"]} {
		if sample, ok := primitiveSample(value); ok && safeAPIParameterValue(sample) {
			return sample, true
		}
	}
	schema := mapField(param, "schema")
	for _, value := range []any{schema["example"], schema["default"]} {
		if sample, ok := primitiveSample(value); ok && safeAPIParameterValue(sample) {
			return sample, true
		}
	}
	for _, value := range sliceField(schema, "enum") {
		if sample, ok := primitiveSample(value); ok && safeAPIParameterValue(sample) {
			return sample, true
		}
	}
	return "", false
}

func primitiveSample(value any) (string, bool) {
	switch typed := value.(type) {
	case nil:
		return "", false
	case string:
		return strings.TrimSpace(typed), strings.TrimSpace(typed) != ""
	case json.Number:
		return typed.String(), typed.String() != ""
	case int:
		return fmt.Sprintf("%d", typed), true
	case int64:
		return fmt.Sprintf("%d", typed), true
	case float64:
		return fmt.Sprintf("%v", typed), true
	case bool:
		if typed {
			return "true", true
		}
		return "false", true
	default:
		return "", false
	}
}

func safeAPIParameterValue(value string) bool {
	if value == "" || len(value) > 80 {
		return false
	}
	lower := strings.ToLower(value)
	if strings.ContainsAny(value, "\r\n") || strings.Contains(value, "/") || strings.Contains(value, "\\") {
		return false
	}
	for _, marker := range []string{"bearer ", "basic ", "token", "password", "secret", "apikey", "api_key", "cookie"} {
		if strings.Contains(lower, marker) {
			return false
		}
	}
	return true
}

func sensitiveAPIParameterName(name string) bool {
	name = strings.ToLower(name)
	for _, marker := range []string{"authorization", "password", "passwd", "token", "secret", "api_key", "apikey", "cookie", "session"} {
		if strings.Contains(name, marker) {
			return true
		}
	}
	return false
}

func expectedStatuses(operation map[string]any) []string {
	responses := mapField(operation, "responses")
	statuses := make([]string, 0, len(responses))
	for status := range responses {
		statuses = append(statuses, status)
	}
	sort.Strings(statuses)
	return statuses
}

func expectedContentTypes(operation map[string]any) []string {
	responses := mapField(operation, "responses")
	seen := make(map[string]struct{})
	for _, rawResponse := range responses {
		response := asMap(rawResponse)
		if response == nil {
			continue
		}
		content := mapField(response, "content")
		for contentType := range content {
			seen[contentType] = struct{}{}
		}
	}
	contentTypes := make([]string, 0, len(seen))
	for contentType := range seen {
		contentTypes = append(contentTypes, contentType)
	}
	sort.Strings(contentTypes)
	return contentTypes
}

func expectedResponseSchemas(operation map[string]any, doc map[string]any) map[string]any {
	responses := mapField(operation, "responses")
	out := make(map[string]any)
	for status, rawResponse := range responses {
		response := asMap(rawResponse)
		if response == nil {
			continue
		}
		content := mapField(response, "content")
		if len(content) == 0 {
			continue
		}
		statusSchemas := make(map[string]any)
		for contentType, rawMedia := range content {
			media := asMap(rawMedia)
			schema := mapField(media, "schema")
			if len(schema) == 0 {
				continue
			}
			statusSchemas[contentType] = resolveLocalSchemaRefs(schema, doc, 0)
		}
		if len(statusSchemas) > 0 {
			out[status] = statusSchemas
		}
	}
	return out
}

func resolveLocalSchemaRefs(schema map[string]any, doc map[string]any, depth int) map[string]any {
	if schema == nil || depth > 6 {
		return map[string]any{}
	}
	if ref := stringField(schema, "$ref"); strings.HasPrefix(ref, "#/") {
		if resolved := resolveJSONPointer(doc, strings.TrimPrefix(ref, "#/")); resolved != nil {
			return resolveLocalSchemaRefs(resolved, doc, depth+1)
		}
	}
	out := make(map[string]any, len(schema))
	for key, value := range schema {
		if key == "example" || key == "examples" || key == "description" {
			continue
		}
		switch typed := value.(type) {
		case map[string]any:
			out[key] = resolveLocalSchemaRefs(typed, doc, depth+1)
		case []any:
			items := make([]any, 0, len(typed))
			for _, item := range typed {
				if itemMap := asMap(item); itemMap != nil {
					items = append(items, resolveLocalSchemaRefs(itemMap, doc, depth+1))
				} else {
					items = append(items, normalizeValue(item))
				}
			}
			out[key] = items
		default:
			out[key] = normalizeValue(value)
		}
	}
	return out
}

func resolveJSONPointer(doc map[string]any, pointer string) map[string]any {
	var current any = doc
	for _, part := range strings.Split(pointer, "/") {
		part = strings.ReplaceAll(strings.ReplaceAll(part, "~1", "/"), "~0", "~")
		currentMap := asMap(current)
		if currentMap == nil {
			return nil
		}
		current = currentMap[part]
	}
	return asMap(current)
}

func resolveServerURL(server string, sourceURL string, apiBaseURL string) string {
	server = strings.TrimSpace(server)
	if server == "" {
		if apiBaseURL != "" {
			return apiBaseURL
		}
		if sourceURL != "" {
			parsed, err := url.Parse(sourceURL)
			if err == nil && parsed.Scheme != "" && parsed.Host != "" {
				parsed.Path = ""
				parsed.RawQuery = ""
				parsed.Fragment = ""
				return parsed.String()
			}
		}
		return ""
	}
	if parsed, err := url.Parse(server); err == nil {
		if parsed.IsAbs() {
			return parsed.String()
		}
		if apiBaseURL != "" {
			base, baseErr := url.Parse(apiBaseURL)
			if baseErr == nil {
				return base.ResolveReference(parsed).String()
			}
		}
		if sourceURL != "" {
			base, baseErr := url.Parse(sourceURL)
			if baseErr == nil {
				return base.ResolveReference(parsed).String()
			}
		}
	}
	return server
}

func contentTypeMatches(actual string, expected []string) bool {
	if len(expected) == 0 || actual == "" {
		return true
	}
	actualMediaType, _, err := mime.ParseMediaType(actual)
	if err != nil {
		actualMediaType = strings.ToLower(strings.TrimSpace(strings.Split(actual, ";")[0]))
	}
	for _, item := range expected {
		expectedMediaType, _, err := mime.ParseMediaType(item)
		if err != nil {
			expectedMediaType = strings.ToLower(strings.TrimSpace(strings.Split(item, ";")[0]))
		}
		if strings.EqualFold(actualMediaType, expectedMediaType) {
			return true
		}
		if strings.HasSuffix(expectedMediaType, "/*") && strings.HasPrefix(actualMediaType, strings.TrimSuffix(expectedMediaType, "*")) {
			return true
		}
		if strings.HasSuffix(expectedMediaType, "+json") && strings.HasSuffix(actualMediaType, "+json") {
			return true
		}
	}
	return false
}

func statusMatchesExpected(status int, expected []string) bool {
	if len(expected) == 0 {
		return true
	}
	value := fmt.Sprintf("%d", status)
	for _, item := range expected {
		if item == value || strings.EqualFold(item, "default") {
			return true
		}
		if len(item) == 3 && strings.HasSuffix(strings.ToUpper(item), "XX") && item[:1] == value[:1] {
			return true
		}
	}
	return false
}

func normalizeMap(input map[string]any) map[string]any {
	out := make(map[string]any, len(input))
	for key, value := range input {
		out[key] = normalizeValue(value)
	}
	return out
}

func normalizeValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		return normalizeMap(typed)
	case map[any]any:
		out := make(map[string]any, len(typed))
		for key, value := range typed {
			out[fmt.Sprint(key)] = normalizeValue(value)
		}
		return out
	case []any:
		for index, item := range typed {
			typed[index] = normalizeValue(item)
		}
		return typed
	default:
		return typed
	}
}

func mapField(input map[string]any, key string) map[string]any {
	return asMap(input[key])
}

func sliceField(input map[string]any, key string) []any {
	return asSlice(input[key])
}

func stringField(input map[string]any, key string) string {
	if input == nil {
		return ""
	}
	value, _ := primitiveSample(input[key])
	return value
}

func boolField(input map[string]any, key string) bool {
	if input == nil {
		return false
	}
	value, ok := input[key].(bool)
	return ok && value
}

func asMap(value any) map[string]any {
	switch typed := value.(type) {
	case map[string]any:
		return typed
	case map[any]any:
		out := make(map[string]any, len(typed))
		for key, value := range typed {
			out[fmt.Sprint(key)] = normalizeValue(value)
		}
		return out
	default:
		return nil
	}
}

func asSlice(value any) []any {
	items, ok := value.([]any)
	if !ok {
		return nil
	}
	return items
}

func firstMap(items []any) map[string]any {
	if len(items) == 0 {
		return nil
	}
	return asMap(items[0])
}

func stringSlice(items []any) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		value, ok := primitiveSample(item)
		if ok && value != "" {
			out = append(out, value)
		}
	}
	return out
}
