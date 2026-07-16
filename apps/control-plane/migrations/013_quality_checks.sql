CREATE TABLE IF NOT EXISTS quality_check_runs (
	id uuid PRIMARY KEY,
	project_id uuid NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
	discovery_run_id uuid REFERENCES discovery_runs(id) ON DELETE SET NULL,
	credential_profile_id uuid REFERENCES credential_profiles(id) ON DELETE SET NULL,
	status text NOT NULL,
	target_url text NOT NULL,
	max_pages integer NOT NULL DEFAULT 10,
	include_security boolean NOT NULL DEFAULT true,
	include_accessibility boolean NOT NULL DEFAULT true,
	include_performance boolean NOT NULL DEFAULT true,
	started_at timestamptz,
	completed_at timestamptz,
	total_pages integer NOT NULL DEFAULT 0,
	total_findings integer NOT NULL DEFAULT 0,
	critical_findings integer NOT NULL DEFAULT 0,
	high_findings integer NOT NULL DEFAULT 0,
	medium_findings integer NOT NULL DEFAULT 0,
	low_findings integer NOT NULL DEFAULT 0,
	info_findings integer NOT NULL DEFAULT 0,
	error_message text NOT NULL DEFAULT '',
	summary_json jsonb NOT NULL DEFAULT '{}'::jsonb,
	created_at timestamptz NOT NULL DEFAULT now(),
	updated_at timestamptz NOT NULL DEFAULT now(),
	CONSTRAINT quality_check_runs_status_check CHECK (status IN ('queued', 'running', 'completed', 'failed', 'error')),
	CONSTRAINT quality_check_runs_max_pages_check CHECK (max_pages >= 1 AND max_pages <= 50)
);

CREATE TABLE IF NOT EXISTS quality_check_results (
	id uuid PRIMARY KEY,
	run_id uuid NOT NULL REFERENCES quality_check_runs(id) ON DELETE CASCADE,
	project_id uuid NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
	category text NOT NULL,
	rule_id text NOT NULL,
	severity text NOT NULL,
	title text NOT NULL,
	description text NOT NULL DEFAULT '',
	recommendation text NOT NULL DEFAULT '',
	url text NOT NULL DEFAULT '',
	evidence_json jsonb NOT NULL DEFAULT '{}'::jsonb,
	created_at timestamptz NOT NULL DEFAULT now(),
	CONSTRAINT quality_check_results_category_check CHECK (category IN ('security', 'accessibility', 'performance')),
	CONSTRAINT quality_check_results_severity_check CHECK (severity IN ('critical', 'high', 'medium', 'low', 'info'))
);

ALTER TABLE qa_runs
	ADD COLUMN IF NOT EXISTS quality_check_run_id uuid REFERENCES quality_check_runs(id) ON DELETE SET NULL;

ALTER TABLE qa_runs DROP CONSTRAINT IF EXISTS qa_runs_status_check;
ALTER TABLE qa_runs
	ADD CONSTRAINT qa_runs_status_check CHECK (status IN (
		'queued',
		'running_discovery',
		'running_quality_checks',
		'generating_plan',
		'previewing_execution',
		'executing_plan',
		'completed',
		'failed',
		'error'
	));

CREATE INDEX IF NOT EXISTS idx_quality_check_runs_project_created_at ON quality_check_runs(project_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_quality_check_runs_status ON quality_check_runs(status);
CREATE INDEX IF NOT EXISTS idx_quality_check_runs_discovery_run_id ON quality_check_runs(discovery_run_id);
CREATE INDEX IF NOT EXISTS idx_quality_check_results_run_id ON quality_check_results(run_id);
CREATE INDEX IF NOT EXISTS idx_quality_check_results_project_id ON quality_check_results(project_id);
CREATE INDEX IF NOT EXISTS idx_quality_check_results_category ON quality_check_results(category);
CREATE INDEX IF NOT EXISTS idx_qa_runs_quality_check_run_id ON qa_runs(quality_check_run_id);
