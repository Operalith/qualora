package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *Store) CreateAPISpec(ctx context.Context, projectID string, input APISpecImportRequest, rawSpec string, parsed *parsedOpenAPISpec, status string, errorMessage string) (*APISpecDetail, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin create api spec: %w", err)
	}
	defer tx.Rollback(ctx)

	spec := APISpec{}
	var (
		title      string
		version    string
		serverURL  string
		operations []APIOperation
	)
	if parsed != nil {
		title = parsed.Title
		version = parsed.Version
		serverURL = parsed.ServerURL
		operations = parsed.Operations
	}
	operationCount, safeCount, skippedCount := countOperations(operations)
	err = tx.QueryRow(ctx, `
INSERT INTO api_specs (
	id, project_id, name, source_type, source_url, raw_spec, parsed_title, parsed_version,
	server_url, operation_count, safe_operation_count, skipped_operation_count, status, error_message
) VALUES ($1, $2, $3, $4, NULLIF($5, ''), NULLIF($6, ''), $7, $8, $9, $10, $11, $12, $13, $14)
RETURNING id, project_id, name, source_type, COALESCE(source_url, ''), parsed_title, parsed_version, server_url,
	operation_count, safe_operation_count, skipped_operation_count, status, error_message, created_at, updated_at
`, uuid.NewString(), projectID, input.Name, input.SourceType, input.SourceURL, rawSpec, title, version, serverURL, operationCount, safeCount, skippedCount, status, errorMessage).Scan(
		&spec.ID,
		&spec.ProjectID,
		&spec.Name,
		&spec.SourceType,
		&spec.SourceURL,
		&spec.ParsedTitle,
		&spec.ParsedVersion,
		&spec.ServerURL,
		&spec.OperationCount,
		&spec.SafeOperationCount,
		&spec.SkippedOperationCount,
		&spec.Status,
		&spec.ErrorMessage,
		&spec.CreatedAt,
		&spec.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert api spec: %w", err)
	}

	storedOperations := make([]APIOperation, 0, len(operations))
	for _, operation := range operations {
		operation.APISpecID = spec.ID
		operation.ProjectID = projectID
		stored, err := insertAPIOperation(ctx, tx, operation)
		if err != nil {
			return nil, err
		}
		storedOperations = append(storedOperations, stored)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit create api spec: %w", err)
	}
	return &APISpecDetail{Spec: spec, Operations: storedOperations}, nil
}

func (s *Store) ListAPISpecs(ctx context.Context, projectID string) ([]APISpec, error) {
	rows, err := s.db.Query(ctx, `
SELECT id, project_id, name, source_type, COALESCE(source_url, ''), parsed_title, parsed_version, server_url,
	operation_count, safe_operation_count, skipped_operation_count, status, error_message, created_at, updated_at
FROM api_specs
WHERE project_id = $1
ORDER BY created_at DESC
`, projectID)
	if err != nil {
		return nil, fmt.Errorf("query api specs: %w", err)
	}
	defer rows.Close()

	specs := make([]APISpec, 0)
	for rows.Next() {
		spec, err := scanAPISpec(rows)
		if err != nil {
			return nil, err
		}
		specs = append(specs, spec)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate api specs: %w", err)
	}
	return specs, nil
}

func (s *Store) GetAPISpec(ctx context.Context, id string) (*APISpec, error) {
	spec, err := scanAPISpec(s.db.QueryRow(ctx, `
SELECT id, project_id, name, source_type, COALESCE(source_url, ''), parsed_title, parsed_version, server_url,
	operation_count, safe_operation_count, skipped_operation_count, status, error_message, created_at, updated_at
FROM api_specs
WHERE id = $1
`, id))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &spec, nil
}

func (s *Store) GetAPISpecDetail(ctx context.Context, id string) (*APISpecDetail, error) {
	spec, err := s.GetAPISpec(ctx, id)
	if err != nil {
		return nil, err
	}
	operations, err := s.ListAPIOperations(ctx, id)
	if err != nil {
		return nil, err
	}
	return &APISpecDetail{Spec: *spec, Operations: operations}, nil
}

func (s *Store) ListAPIOperations(ctx context.Context, apiSpecID string) ([]APIOperation, error) {
	rows, err := s.db.Query(ctx, `
SELECT id, api_spec_id, project_id, method, path, resolved_path, query_string, operation_id, summary, description,
	tags_json, expected_statuses_json, expected_content_types_json, response_schemas_json, requires_authentication, safe_to_execute, skip_reason,
	created_at, updated_at
FROM api_operations
WHERE api_spec_id = $1
ORDER BY path ASC, method ASC
`, apiSpecID)
	if err != nil {
		return nil, fmt.Errorf("query api operations: %w", err)
	}
	defer rows.Close()

	operations := make([]APIOperation, 0)
	for rows.Next() {
		operation, err := scanAPIOperation(rows)
		if err != nil {
			return nil, err
		}
		operations = append(operations, operation)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate api operations: %w", err)
	}
	return operations, nil
}

func (s *Store) DeleteAPISpec(ctx context.Context, id string) error {
	tag, err := s.db.Exec(ctx, `DELETE FROM api_specs WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete api spec: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) CreateAPISmokeRunRecord(ctx context.Context, projectID string, apiSpecID string, apiAuthProfileID string) (*TestRun, error) {
	run := TestRun{}
	var specID sql.NullString
	var credentialProfileID sql.NullString
	var storedAPIAuthProfileID sql.NullString
	err := s.db.QueryRow(ctx, `
INSERT INTO test_runs (id, project_id, run_type, api_spec_id, api_auth_profile_id, status, started_at)
VALUES ($1, $2, $3, $4, NULLIF($5, '')::uuid, $6, now())
RETURNING id, project_id, run_type, api_spec_id::text, credential_profile_id::text, api_auth_profile_id::text,
	target_path, capture_screenshot, max_duration_seconds, status, error_message,
	page_title, started_at, completed_at, created_at, updated_at
`, uuid.NewString(), projectID, RunTypeAPISmoke, apiSpecID, apiAuthProfileID, StatusRunning).Scan(
		&run.ID,
		&run.ProjectID,
		&run.RunType,
		&specID,
		&credentialProfileID,
		&storedAPIAuthProfileID,
		&run.TargetPath,
		&run.CaptureScreenshot,
		&run.MaxDurationSeconds,
		&run.Status,
		&run.ErrorMessage,
		&run.PageTitle,
		&run.StartedAt,
		&run.CompletedAt,
		&run.CreatedAt,
		&run.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert api smoke run: %w", err)
	}
	if specID.Valid {
		run.APISpecID = specID.String
	}
	if credentialProfileID.Valid {
		run.CredentialProfileID = credentialProfileID.String
	}
	if storedAPIAuthProfileID.Valid {
		run.APIAuthProfileID = storedAPIAuthProfileID.String
	}
	return &run, nil
}

func (s *Store) CompleteAPISmokeRun(ctx context.Context, runID string, status string, errorMessage string) error {
	if _, err := s.db.Exec(ctx, `
UPDATE test_runs
SET status = $2, error_message = $3, completed_at = now(), updated_at = now()
WHERE id = $1
`, runID, status, errorMessage); err != nil {
		return fmt.Errorf("complete api smoke run: %w", err)
	}
	return nil
}

func (s *Store) InsertAPICheckResult(ctx context.Context, result APICheckResult) (*APICheckResult, error) {
	if result.ID == "" {
		result.ID = uuid.NewString()
	}
	var operationID any
	if result.OperationID != "" {
		operationID = result.OperationID
	}
	var (
		returnedOperationID sql.NullString
		apiAuthProfileID    sql.NullString
		schemaErrorsRaw     []byte
		expectedStatusesRaw []byte
		expectedTypesRaw    []byte
	)
	err := s.db.QueryRow(ctx, `
INSERT INTO api_check_results (
	id, run_id, api_spec_id, operation_id, method, path, resolved_url, status, http_status,
	duration_ms, response_content_type, response_size_bytes, error_message, skipped_reason,
	api_auth_profile_id, auth_mode, contract_validation_status, schema_validation_errors_json,
	expected_statuses_json, actual_status, expected_content_types_json, actual_content_type,
	response_time_ms, unauthenticated_status
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14,
	NULLIF($15, '')::uuid, $16, $17, $18, $19, $20, $21, $22, $23, $24)
RETURNING id, run_id, api_spec_id, operation_id::text, method, path, resolved_url, status, http_status,
	duration_ms, response_content_type, response_size_bytes, error_message, skipped_reason,
	api_auth_profile_id::text, auth_mode, contract_validation_status, schema_validation_errors_json,
	expected_statuses_json, actual_status, expected_content_types_json, actual_content_type,
	response_time_ms, unauthenticated_status, created_at
`,
		result.ID,
		result.RunID,
		result.APISpecID,
		operationID,
		result.Method,
		result.Path,
		result.ResolvedURL,
		result.Status,
		result.HTTPStatus,
		result.DurationMS,
		result.ResponseContentType,
		result.ResponseSizeBytes,
		result.ErrorMessage,
		result.SkippedReason,
		result.APIAuthProfileID,
		firstNonEmpty(result.AuthMode, APIAuthProfileTypeNone),
		firstNonEmpty(result.ContractValidationStatus, "unknown"),
		mustJSON(result.SchemaValidationErrors),
		mustJSON(result.ExpectedStatuses),
		result.ActualStatus,
		mustJSON(result.ExpectedContentTypes),
		result.ActualContentType,
		result.ResponseTimeMS,
		result.UnauthenticatedStatus,
	).Scan(
		&result.ID,
		&result.RunID,
		&result.APISpecID,
		&returnedOperationID,
		&result.Method,
		&result.Path,
		&result.ResolvedURL,
		&result.Status,
		&result.HTTPStatus,
		&result.DurationMS,
		&result.ResponseContentType,
		&result.ResponseSizeBytes,
		&result.ErrorMessage,
		&result.SkippedReason,
		&apiAuthProfileID,
		&result.AuthMode,
		&result.ContractValidationStatus,
		&schemaErrorsRaw,
		&expectedStatusesRaw,
		&result.ActualStatus,
		&expectedTypesRaw,
		&result.ActualContentType,
		&result.ResponseTimeMS,
		&result.UnauthenticatedStatus,
		&result.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert api check result: %w", err)
	}
	if returnedOperationID.Valid {
		result.OperationID = returnedOperationID.String
	}
	if apiAuthProfileID.Valid {
		result.APIAuthProfileID = apiAuthProfileID.String
	}
	result.SchemaValidationErrors = mustStringList(schemaErrorsRaw)
	result.ExpectedStatuses = mustStringList(expectedStatusesRaw)
	result.ExpectedContentTypes = mustStringList(expectedTypesRaw)
	return &result, nil
}

func (s *Store) ListAPICheckResults(ctx context.Context, runID string) ([]APICheckResult, error) {
	rows, err := s.db.Query(ctx, `
SELECT id, run_id, api_spec_id, operation_id::text, method, path, resolved_url, status, http_status,
	duration_ms, response_content_type, response_size_bytes, error_message, skipped_reason,
	api_auth_profile_id::text, auth_mode, contract_validation_status, schema_validation_errors_json,
	expected_statuses_json, actual_status, expected_content_types_json, actual_content_type,
	response_time_ms, unauthenticated_status, created_at
FROM api_check_results
WHERE run_id = $1
ORDER BY created_at ASC
`, runID)
	if err != nil {
		return nil, fmt.Errorf("query api check results: %w", err)
	}
	defer rows.Close()

	results := make([]APICheckResult, 0)
	for rows.Next() {
		result, err := scanAPICheckResult(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate api check results: %w", err)
	}
	return results, nil
}

func (s *Store) InsertRunEvidence(ctx context.Context, runID string, evidence Evidence) (string, error) {
	id := uuid.NewString()
	metadata, err := json.Marshal(evidence.Metadata)
	if err != nil {
		return "", fmt.Errorf("marshal evidence metadata: %w", err)
	}
	if _, err := s.db.Exec(ctx, `
INSERT INTO evidence (id, run_id, type, uri, metadata)
VALUES ($1, $2, $3, $4, $5)
`, id, runID, evidence.Type, evidence.URI, metadata); err != nil {
		return "", fmt.Errorf("insert evidence: %w", err)
	}
	return id, nil
}

func (s *Store) InsertRunFinding(ctx context.Context, runID string, finding Finding) error {
	evidenceIDs, err := json.Marshal(finding.EvidenceIDs)
	if err != nil {
		return fmt.Errorf("marshal finding evidence ids: %w", err)
	}
	if _, err := s.db.Exec(ctx, `
INSERT INTO findings (id, run_id, title, severity, category, confidence, description, recommendation, evidence_ids)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
`, uuid.NewString(), runID, finding.Title, finding.Severity, finding.Category, finding.Confidence, finding.Description, finding.Recommendation, evidenceIDs); err != nil {
		return fmt.Errorf("insert finding: %w", err)
	}
	return nil
}

func insertAPIOperation(ctx context.Context, tx pgx.Tx, operation APIOperation) (APIOperation, error) {
	tags, err := json.Marshal(operation.Tags)
	if err != nil {
		return APIOperation{}, fmt.Errorf("marshal operation tags: %w", err)
	}
	statuses, err := json.Marshal(operation.ExpectedStatuses)
	if err != nil {
		return APIOperation{}, fmt.Errorf("marshal operation statuses: %w", err)
	}
	contentTypes, err := json.Marshal(operation.ExpectedContentTypes)
	if err != nil {
		return APIOperation{}, fmt.Errorf("marshal operation content types: %w", err)
	}
	responseSchemas, err := json.Marshal(operation.ResponseSchemas)
	if err != nil {
		return APIOperation{}, fmt.Errorf("marshal operation response schemas: %w", err)
	}
	var requiresAuthParam any
	if operation.RequiresAuthentication != nil {
		requiresAuthParam = *operation.RequiresAuthentication
	}
	var requiresAuth sql.NullBool
	err = tx.QueryRow(ctx, `
INSERT INTO api_operations (
	id, api_spec_id, project_id, method, path, resolved_path, query_string, operation_id, summary, description,
	tags_json, expected_statuses_json, expected_content_types_json, response_schemas_json, requires_authentication, safe_to_execute, skip_reason
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
RETURNING id, api_spec_id, project_id, method, path, resolved_path, query_string, operation_id, summary, description,
	tags_json, expected_statuses_json, expected_content_types_json, response_schemas_json, requires_authentication, safe_to_execute, skip_reason,
	created_at, updated_at
`, uuid.NewString(), operation.APISpecID, operation.ProjectID, operation.Method, operation.Path, operation.ResolvedPath, operation.QueryString,
		operation.OperationID, operation.Summary, operation.Description, tags, statuses, contentTypes, responseSchemas, requiresAuthParam, operation.SafeToExecute, operation.SkipReason).Scan(
		&operation.ID,
		&operation.APISpecID,
		&operation.ProjectID,
		&operation.Method,
		&operation.Path,
		&operation.ResolvedPath,
		&operation.QueryString,
		&operation.OperationID,
		&operation.Summary,
		&operation.Description,
		&tags,
		&statuses,
		&contentTypes,
		&responseSchemas,
		&requiresAuth,
		&operation.SafeToExecute,
		&operation.SkipReason,
		&operation.CreatedAt,
		&operation.UpdatedAt,
	)
	if err != nil {
		return APIOperation{}, fmt.Errorf("insert api operation: %w", err)
	}
	operation.Tags = mustStringList(tags)
	operation.ExpectedStatuses = mustStringList(statuses)
	operation.ExpectedContentTypes = mustStringList(contentTypes)
	operation.ResponseSchemas = mustMap(responseSchemas)
	if requiresAuth.Valid {
		value := requiresAuth.Bool
		operation.RequiresAuthentication = &value
	}
	return operation, nil
}

func scanAPISpec(row scanRow) (APISpec, error) {
	var spec APISpec
	if err := row.Scan(
		&spec.ID,
		&spec.ProjectID,
		&spec.Name,
		&spec.SourceType,
		&spec.SourceURL,
		&spec.ParsedTitle,
		&spec.ParsedVersion,
		&spec.ServerURL,
		&spec.OperationCount,
		&spec.SafeOperationCount,
		&spec.SkippedOperationCount,
		&spec.Status,
		&spec.ErrorMessage,
		&spec.CreatedAt,
		&spec.UpdatedAt,
	); err != nil {
		return APISpec{}, fmt.Errorf("scan api spec: %w", err)
	}
	return spec, nil
}

func scanAPIOperation(row scanRow) (APIOperation, error) {
	var (
		operation       APIOperation
		tags            []byte
		statuses        []byte
		contentTypes    []byte
		responseSchemas []byte
		requiresAuth    sql.NullBool
	)
	if err := row.Scan(
		&operation.ID,
		&operation.APISpecID,
		&operation.ProjectID,
		&operation.Method,
		&operation.Path,
		&operation.ResolvedPath,
		&operation.QueryString,
		&operation.OperationID,
		&operation.Summary,
		&operation.Description,
		&tags,
		&statuses,
		&contentTypes,
		&responseSchemas,
		&requiresAuth,
		&operation.SafeToExecute,
		&operation.SkipReason,
		&operation.CreatedAt,
		&operation.UpdatedAt,
	); err != nil {
		return APIOperation{}, fmt.Errorf("scan api operation: %w", err)
	}
	operation.Tags = mustStringList(tags)
	operation.ExpectedStatuses = mustStringList(statuses)
	operation.ExpectedContentTypes = mustStringList(contentTypes)
	operation.ResponseSchemas = mustMap(responseSchemas)
	if requiresAuth.Valid {
		value := requiresAuth.Bool
		operation.RequiresAuthentication = &value
	}
	return operation, nil
}

func scanAPICheckResult(row scanRow) (APICheckResult, error) {
	var (
		result                APICheckResult
		operationID           sql.NullString
		apiAuthProfileID      sql.NullString
		httpStatus            sql.NullInt32
		durationMS            sql.NullInt32
		responseBytes         sql.NullInt32
		schemaErrorsRaw       []byte
		expectedStatusesRaw   []byte
		actualStatus          sql.NullInt32
		expectedTypesRaw      []byte
		responseTimeMS        sql.NullInt32
		unauthenticatedStatus sql.NullInt32
	)
	if err := row.Scan(
		&result.ID,
		&result.RunID,
		&result.APISpecID,
		&operationID,
		&result.Method,
		&result.Path,
		&result.ResolvedURL,
		&result.Status,
		&httpStatus,
		&durationMS,
		&result.ResponseContentType,
		&responseBytes,
		&result.ErrorMessage,
		&result.SkippedReason,
		&apiAuthProfileID,
		&result.AuthMode,
		&result.ContractValidationStatus,
		&schemaErrorsRaw,
		&expectedStatusesRaw,
		&actualStatus,
		&expectedTypesRaw,
		&result.ActualContentType,
		&responseTimeMS,
		&unauthenticatedStatus,
		&result.CreatedAt,
	); err != nil {
		return APICheckResult{}, fmt.Errorf("scan api check result: %w", err)
	}
	if operationID.Valid {
		result.OperationID = operationID.String
	}
	if apiAuthProfileID.Valid {
		result.APIAuthProfileID = apiAuthProfileID.String
	}
	if httpStatus.Valid {
		value := int(httpStatus.Int32)
		result.HTTPStatus = &value
	}
	if durationMS.Valid {
		value := int(durationMS.Int32)
		result.DurationMS = &value
	}
	if actualStatus.Valid {
		value := int(actualStatus.Int32)
		result.ActualStatus = &value
	}
	if responseTimeMS.Valid {
		value := int(responseTimeMS.Int32)
		result.ResponseTimeMS = &value
	}
	if unauthenticatedStatus.Valid {
		value := int(unauthenticatedStatus.Int32)
		result.UnauthenticatedStatus = &value
	}
	if responseBytes.Valid {
		value := int(responseBytes.Int32)
		result.ResponseSizeBytes = &value
	}
	result.SchemaValidationErrors = mustStringList(schemaErrorsRaw)
	result.ExpectedStatuses = mustStringList(expectedStatusesRaw)
	result.ExpectedContentTypes = mustStringList(expectedTypesRaw)
	return result, nil
}

func countOperations(operations []APIOperation) (int, int, int) {
	total := len(operations)
	safe := 0
	for _, operation := range operations {
		if operation.SafeToExecute {
			safe++
		}
	}
	return total, safe, total - safe
}

func summarizeAPICheckResults(results []APICheckResult) APISmokeSummary {
	summary := APISmokeSummary{TotalOperations: len(results)}
	for _, result := range results {
		switch result.Status {
		case StatusPassed:
			summary.ExecutedOperations++
			summary.PassedOperations++
		case StatusFailed:
			summary.ExecutedOperations++
			summary.FailedOperations++
		case StatusError:
			summary.ExecutedOperations++
			summary.ErroredOperations++
		case StatusSkipped:
			summary.SkippedOperations++
		}
		if result.APIAuthProfileID != "" && result.Status != StatusSkipped {
			summary.AuthenticatedOperations++
		}
		if result.UnauthenticatedStatus != nil {
			summary.UnauthenticatedComparisons++
		}
		switch result.ContractValidationStatus {
		case "passed":
			summary.ContractPassed++
		case "failed":
			summary.ContractFailed++
		case "skipped":
			summary.ContractSkipped++
		default:
			summary.ContractUnknown++
		}
		summary.SchemaValidationErrorCount += len(result.SchemaValidationErrors)
	}
	return summary
}

func mustStringList(raw []byte) []string {
	var values []string
	_ = json.Unmarshal(raw, &values)
	if values == nil {
		return []string{}
	}
	return values
}

func mustMap(raw []byte) map[string]any {
	var values map[string]any
	_ = json.Unmarshal(raw, &values)
	if values == nil {
		return map[string]any{}
	}
	return values
}

func mustJSON(value any) []byte {
	raw, err := json.Marshal(value)
	if err != nil || raw == nil {
		return []byte("null")
	}
	return raw
}
