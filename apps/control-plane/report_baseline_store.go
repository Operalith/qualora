package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *Store) GetReportSnapshot(ctx context.Context, reportType string, reportID string) (ReportSnapshot, error) {
	normalizedType, err := NormalizeReportType(reportType)
	if err != nil {
		return ReportSnapshot{}, err
	}
	switch normalizedType {
	case ReportTypeSafeQA:
		report, err := s.GetQARunReport(ctx, reportID)
		if err != nil {
			return ReportSnapshot{}, err
		}
		return ReportSnapshot{
			ProjectID:    report.Project.ID,
			ReportType:   ReportTypeSafeQA,
			ReportID:     report.Run.ID,
			SourceRunID:  report.Run.ID,
			Status:       report.Run.Status,
			Intelligence: report.ReportIntelligence,
		}, nil
	case ReportTypeQualityCheck:
		report, err := s.GetQualityCheckReport(ctx, reportID)
		if err != nil {
			return ReportSnapshot{}, err
		}
		return ReportSnapshot{
			ProjectID:    report.Project.ID,
			ReportType:   ReportTypeQualityCheck,
			ReportID:     report.Run.ID,
			SourceRunID:  report.Run.ID,
			Status:       report.Run.Status,
			Intelligence: report.ReportIntelligence,
		}, nil
	case ReportTypeDiscovery:
		report, err := s.GetDiscoveryReport(ctx, reportID)
		if err != nil {
			return ReportSnapshot{}, err
		}
		return ReportSnapshot{
			ProjectID:    report.Project.ID,
			ReportType:   ReportTypeDiscovery,
			ReportID:     report.Run.ID,
			SourceRunID:  report.Run.ID,
			Status:       report.Run.Status,
			Intelligence: report.ReportIntelligence,
		}, nil
	case ReportTypeSafeExplorer:
		report, err := s.GetSafeExplorerReport(ctx, reportID)
		if err != nil {
			return ReportSnapshot{}, err
		}
		return ReportSnapshot{
			ProjectID:    report.Project.ID,
			ReportType:   ReportTypeSafeExplorer,
			ReportID:     report.Run.ID,
			SourceRunID:  report.Run.ID,
			Status:       report.Run.Status,
			Intelligence: report.ReportIntelligence,
		}, nil
	case ReportTypeAuthorization:
		report, err := s.GetAuthorizationCheckReport(ctx, reportID)
		if err != nil {
			return ReportSnapshot{}, err
		}
		return ReportSnapshot{
			ProjectID:    report.Project.ID,
			ReportType:   ReportTypeAuthorization,
			ReportID:     report.Run.ID,
			SourceRunID:  report.Run.ID,
			Status:       report.Run.Status,
			Intelligence: report.ReportIntelligence,
		}, nil
	case ReportTypeAPISmoke, ReportTypeBrowserSmoke:
		report, err := s.GetReport(ctx, reportID)
		if err != nil {
			return ReportSnapshot{}, err
		}
		expectedRunType := RunTypeAPISmoke
		if normalizedType == ReportTypeBrowserSmoke {
			expectedRunType = RunTypeBrowserSmoke
		}
		if report.RunType != expectedRunType {
			return ReportSnapshot{}, fmt.Errorf("report %s has run_type %q, expected %q", reportID, report.RunType, expectedRunType)
		}
		return ReportSnapshot{
			ProjectID:    report.ProjectID,
			ReportType:   normalizedType,
			ReportID:     report.RunID,
			SourceRunID:  report.RunID,
			Status:       report.Status,
			Intelligence: report.ReportIntelligence,
		}, nil
	default:
		return ReportSnapshot{}, fmt.Errorf("unsupported report_type %q", reportType)
	}
}

func (s *Store) CreateReportBaseline(ctx context.Context, baseline ReportBaseline) (*ReportBaseline, error) {
	fingerprintSet, err := json.Marshal(baseline.FingerprintSet)
	if err != nil {
		return nil, fmt.Errorf("marshal baseline fingerprint set: %w", err)
	}
	severityCounts, err := json.Marshal(baseline.SeverityCounts)
	if err != nil {
		return nil, fmt.Errorf("marshal baseline severity counts: %w", err)
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin create report baseline: %w", err)
	}
	defer tx.Rollback(ctx)

	if baseline.IsDefault {
		if _, err := tx.Exec(ctx, `
UPDATE report_baselines
SET is_default = false, updated_at = now()
WHERE project_id = $1 AND report_type = $2 AND is_default = true
`, baseline.ProjectID, baseline.ReportType); err != nil {
			return nil, fmt.Errorf("unset existing default baseline: %w", err)
		}
	}

	if baseline.ID == "" {
		baseline.ID = uuid.NewString()
	}
	row := tx.QueryRow(ctx, `
INSERT INTO report_baselines (
	id, project_id, name, description, report_type, report_id, source_run_id,
	fingerprint_set_json, severity_counts_json, grouped_findings_count,
	raw_findings_count, created_by_user_id, is_default
) VALUES (
	$1, $2, $3, NULLIF($4, ''), $5, $6, NULLIF($7, ''),
	$8, $9, $10, $11, NULLIF($12, '')::uuid, $13
)
RETURNING id, project_id, name, description, report_type, report_id, source_run_id,
	fingerprint_set_json, severity_counts_json, grouped_findings_count,
	raw_findings_count, created_by_user_id::text, is_default, created_at, updated_at
`, baseline.ID, baseline.ProjectID, baseline.Name, baseline.Description, baseline.ReportType, baseline.ReportID, baseline.SourceRunID, fingerprintSet, severityCounts, baseline.GroupedFindingsCount, baseline.RawFindingsCount, baseline.CreatedByUserID, baseline.IsDefault)

	created, err := scanReportBaseline(row)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit create report baseline: %w", err)
	}
	return &created, nil
}

func (s *Store) ListReportBaselines(ctx context.Context, projectID string, reportType string) ([]ReportBaseline, error) {
	args := []any{projectID}
	filter := ""
	if reportType != "" {
		args = append(args, reportType)
		filter = "AND report_type = $2"
	}
	rows, err := s.db.Query(ctx, `
SELECT id, project_id, name, description, report_type, report_id, source_run_id,
	fingerprint_set_json, severity_counts_json, grouped_findings_count,
	raw_findings_count, created_by_user_id::text, is_default, created_at, updated_at
FROM report_baselines
WHERE project_id = $1 `+filter+`
ORDER BY is_default DESC, created_at DESC
`, args...)
	if err != nil {
		return nil, fmt.Errorf("query report baselines: %w", err)
	}
	defer rows.Close()

	baselines := []ReportBaseline{}
	for rows.Next() {
		baseline, err := scanReportBaseline(rows)
		if err != nil {
			return nil, err
		}
		baselines = append(baselines, baseline)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate report baselines: %w", err)
	}
	return baselines, nil
}

func (s *Store) GetReportBaseline(ctx context.Context, id string) (*ReportBaseline, error) {
	baseline, err := scanReportBaseline(s.db.QueryRow(ctx, `
SELECT id, project_id, name, description, report_type, report_id, source_run_id,
	fingerprint_set_json, severity_counts_json, grouped_findings_count,
	raw_findings_count, created_by_user_id::text, is_default, created_at, updated_at
FROM report_baselines
WHERE id = $1
`, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &baseline, nil
}

func (s *Store) GetDefaultReportBaseline(ctx context.Context, projectID string, reportType string) (*ReportBaseline, error) {
	baseline, err := scanReportBaseline(s.db.QueryRow(ctx, `
SELECT id, project_id, name, description, report_type, report_id, source_run_id,
	fingerprint_set_json, severity_counts_json, grouped_findings_count,
	raw_findings_count, created_by_user_id::text, is_default, created_at, updated_at
FROM report_baselines
WHERE project_id = $1 AND report_type = $2 AND is_default = true
ORDER BY created_at DESC
LIMIT 1
`, projectID, reportType))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &baseline, nil
}

func (s *Store) UpdateReportBaseline(ctx context.Context, id string, input ReportBaselineUpdateRequest) (*ReportBaseline, error) {
	existing, err := s.GetReportBaseline(ctx, id)
	if err != nil {
		return nil, err
	}
	name := existing.Name
	if input.Name != "" {
		name = input.Name
	}
	description := existing.Description
	if input.Description != "" {
		description = input.Description
	}
	isDefault := existing.IsDefault
	if input.IsDefault != nil {
		isDefault = *input.IsDefault
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin update report baseline: %w", err)
	}
	defer tx.Rollback(ctx)

	if isDefault {
		if _, err := tx.Exec(ctx, `
UPDATE report_baselines
SET is_default = false, updated_at = now()
WHERE project_id = $1 AND report_type = $2 AND id <> $3 AND is_default = true
`, existing.ProjectID, existing.ReportType, id); err != nil {
			return nil, fmt.Errorf("unset existing default baseline: %w", err)
		}
	}

	updated, err := scanReportBaseline(tx.QueryRow(ctx, `
UPDATE report_baselines
SET name = $2, description = NULLIF($3, ''), is_default = $4, updated_at = now()
WHERE id = $1
RETURNING id, project_id, name, description, report_type, report_id, source_run_id,
	fingerprint_set_json, severity_counts_json, grouped_findings_count,
	raw_findings_count, created_by_user_id::text, is_default, created_at, updated_at
`, id, name, description, isDefault))
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit update report baseline: %w", err)
	}
	return &updated, nil
}

func (s *Store) DeleteReportBaseline(ctx context.Context, id string) error {
	tag, err := s.db.Exec(ctx, `DELETE FROM report_baselines WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete report baseline: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func scanReportBaseline(row scanRow) (ReportBaseline, error) {
	var baseline ReportBaseline
	var description sql.NullString
	var sourceRunID sql.NullString
	var createdByUserID sql.NullString
	var fingerprintSetRaw []byte
	var severityCountsRaw []byte
	if err := row.Scan(
		&baseline.ID,
		&baseline.ProjectID,
		&baseline.Name,
		&description,
		&baseline.ReportType,
		&baseline.ReportID,
		&sourceRunID,
		&fingerprintSetRaw,
		&severityCountsRaw,
		&baseline.GroupedFindingsCount,
		&baseline.RawFindingsCount,
		&createdByUserID,
		&baseline.IsDefault,
		&baseline.CreatedAt,
		&baseline.UpdatedAt,
	); err != nil {
		return ReportBaseline{}, fmt.Errorf("scan report baseline: %w", err)
	}
	if description.Valid {
		baseline.Description = description.String
	}
	if sourceRunID.Valid {
		baseline.SourceRunID = sourceRunID.String
	}
	if createdByUserID.Valid {
		baseline.CreatedByUserID = createdByUserID.String
	}
	if err := json.Unmarshal(fingerprintSetRaw, &baseline.FingerprintSet); err != nil {
		return ReportBaseline{}, fmt.Errorf("unmarshal report baseline fingerprint set: %w", err)
	}
	if err := json.Unmarshal(severityCountsRaw, &baseline.SeverityCounts); err != nil {
		return ReportBaseline{}, fmt.Errorf("unmarshal report baseline severity counts: %w", err)
	}
	return baseline, nil
}
