CREATE TABLE IF NOT EXISTS ci_runs (
	id UUID PRIMARY KEY,
	project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
	qa_run_id UUID REFERENCES qa_runs(id) ON DELETE SET NULL,
	baseline_id UUID REFERENCES report_baselines(id) ON DELETE SET NULL,
	status TEXT NOT NULL CHECK (status IN ('passed', 'failed', 'warning', 'running', 'error')),
	exit_code INTEGER NOT NULL DEFAULT 1,
	gate_status TEXT NOT NULL DEFAULT '',
	comparison_status TEXT NOT NULL DEFAULT '',
	report_url TEXT NOT NULL DEFAULT '',
	html_report_url TEXT NOT NULL DEFAULT '',
	issue_export_status TEXT NOT NULL DEFAULT '',
	summary_json JSONB NOT NULL DEFAULT '{}'::jsonb,
	started_at TIMESTAMPTZ,
	completed_at TIMESTAMPTZ,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	error_message TEXT NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_ci_runs_project_created_at ON ci_runs(project_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_ci_runs_qa_run_id ON ci_runs(qa_run_id);

CREATE TABLE IF NOT EXISTS issue_export_configs (
	id UUID PRIMARY KEY,
	project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
	provider TEXT NOT NULL CHECK (provider IN ('github', 'gitlab')),
	name TEXT NOT NULL,
	base_url TEXT NOT NULL DEFAULT '',
	owner_or_namespace TEXT NOT NULL,
	repository_or_project TEXT NOT NULL,
	token_encrypted TEXT NOT NULL DEFAULT '',
	default_labels_json JSONB NOT NULL DEFAULT '[]'::jsonb,
	enabled BOOLEAN NOT NULL DEFAULT true,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_issue_export_configs_project_created_at ON issue_export_configs(project_id, created_at DESC);
