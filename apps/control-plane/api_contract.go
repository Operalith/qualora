package main

import (
	"encoding/json"
	"fmt"
	"mime"
	"strings"
)

func validateAPISchemaFindings(operation APIOperation, result APICheckResult, body []byte) []Finding {
	schema := selectAPIResponseSchema(operation, result)
	if len(schema) == 0 {
		return nil
	}
	var payload any
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil
	}
	errors := validateValueAgainstSchema(payload, schema, "$", 0)
	if len(errors) == 0 {
		return nil
	}
	title := "API contract schema mismatch"
	category := "api_contract_schema_mismatch"
	for _, item := range errors {
		if strings.Contains(item, "missing required field") {
			title = "API contract required field missing"
			category = "api_contract_required_field_missing"
			break
		}
	}
	return []Finding{
		{
			Title:          title,
			Severity:       "medium",
			Category:       category,
			Confidence:     "medium",
			Description:    fmt.Sprintf("%s %s response did not match the documented OpenAPI schema: %s", result.Method, result.Path, strings.Join(limitStrings(errors, 5), "; ")),
			Recommendation: "Align the API response shape with the OpenAPI schema or update the schema to match the actual safe response.",
		},
	}
}

func selectAPIResponseSchema(operation APIOperation, result APICheckResult) map[string]any {
	if len(operation.ResponseSchemas) == 0 || result.HTTPStatus == nil {
		return nil
	}
	statusKeys := []string{fmt.Sprintf("%d", *result.HTTPStatus), fmt.Sprintf("%dXX", *result.HTTPStatus/100), fmt.Sprintf("%dxx", *result.HTTPStatus/100), "default"}
	for _, statusKey := range statusKeys {
		rawByContentType, ok := operation.ResponseSchemas[statusKey]
		if !ok {
			continue
		}
		byContentType := asMap(rawByContentType)
		if len(byContentType) == 0 {
			continue
		}
		if schema := schemaForContentType(byContentType, result.ResponseContentType); schema != nil {
			return schema
		}
	}
	return nil
}

func schemaForContentType(byContentType map[string]any, actualContentType string) map[string]any {
	if len(byContentType) == 0 {
		return nil
	}
	actualMediaType := strings.ToLower(strings.TrimSpace(strings.Split(actualContentType, ";")[0]))
	if parsed, _, err := mime.ParseMediaType(actualContentType); err == nil {
		actualMediaType = strings.ToLower(parsed)
	}
	for expected, rawSchema := range byContentType {
		if contentTypeMatches(actualMediaType, []string{expected}) {
			return asMap(rawSchema)
		}
	}
	if actualMediaType == "" {
		for _, rawSchema := range byContentType {
			return asMap(rawSchema)
		}
	}
	return nil
}

func validateValueAgainstSchema(value any, schema map[string]any, path string, depth int) []string {
	if schema == nil || depth > 6 {
		return nil
	}
	if value == nil {
		if schemaAllowsNull(schema) {
			return nil
		}
		return []string{path + " must not be null"}
	}
	errors := make([]string, 0)
	schemaType := schemaPrimaryType(schema)
	if schemaType != "" && !valueMatchesSchemaType(value, schemaType) {
		errors = append(errors, fmt.Sprintf("%s expected %s", path, schemaType))
		return errors
	}
	if enum := sliceField(schema, "enum"); len(enum) > 0 && !valueMatchesEnum(value, enum) {
		errors = append(errors, fmt.Sprintf("%s did not match documented enum", path))
	}
	switch schemaType {
	case "object":
		object, ok := value.(map[string]any)
		if !ok {
			errors = append(errors, fmt.Sprintf("%s expected object", path))
			return errors
		}
		for _, required := range stringSlice(sliceField(schema, "required")) {
			if _, ok := object[required]; !ok {
				errors = append(errors, fmt.Sprintf("%s missing required field %s", path, required))
			}
		}
		properties := mapField(schema, "properties")
		for name, rawPropertySchema := range properties {
			propertySchema := asMap(rawPropertySchema)
			if propertySchema == nil {
				continue
			}
			if propertyValue, ok := object[name]; ok {
				errors = append(errors, validateValueAgainstSchema(propertyValue, propertySchema, path+"."+name, depth+1)...)
			}
		}
	case "array":
		items, ok := value.([]any)
		if !ok {
			errors = append(errors, fmt.Sprintf("%s expected array", path))
			return errors
		}
		itemSchema := mapField(schema, "items")
		if len(itemSchema) > 0 {
			for index, item := range items {
				errors = append(errors, validateValueAgainstSchema(item, itemSchema, fmt.Sprintf("%s[%d]", path, index), depth+1)...)
				if len(errors) >= 20 {
					return errors
				}
			}
		}
	}
	return errors
}

func schemaAllowsNull(schema map[string]any) bool {
	if boolField(schema, "nullable") {
		return true
	}
	for _, item := range sliceField(schema, "type") {
		if value, ok := primitiveSample(item); ok && value == "null" {
			return true
		}
	}
	return false
}

func schemaPrimaryType(schema map[string]any) string {
	if value := stringField(schema, "type"); value != "" {
		return value
	}
	for _, item := range sliceField(schema, "type") {
		if value, ok := primitiveSample(item); ok && value != "null" {
			return value
		}
	}
	if len(mapField(schema, "properties")) > 0 || len(sliceField(schema, "required")) > 0 {
		return "object"
	}
	if len(mapField(schema, "items")) > 0 {
		return "array"
	}
	return ""
}

func valueMatchesSchemaType(value any, schemaType string) bool {
	switch schemaType {
	case "object":
		_, ok := value.(map[string]any)
		return ok
	case "array":
		_, ok := value.([]any)
		return ok
	case "string":
		_, ok := value.(string)
		return ok
	case "number":
		_, ok := value.(float64)
		return ok
	case "integer":
		number, ok := value.(float64)
		return ok && number == float64(int64(number))
	case "boolean":
		_, ok := value.(bool)
		return ok
	default:
		return true
	}
}

func valueMatchesEnum(value any, enum []any) bool {
	actual := fmt.Sprint(value)
	for _, item := range enum {
		if fmt.Sprint(item) == actual {
			return true
		}
	}
	return false
}

func limitStrings(items []string, limit int) []string {
	if len(items) <= limit {
		return items
	}
	return items[:limit]
}
