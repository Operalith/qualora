CREATE TABLE IF NOT EXISTS form_test_runs (
	id uuid PRIMARY KEY,
	project_id uuid NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
	discovery_run_id uuid REFERENCES discovery_runs(id) ON DELETE SET NULL,
	credential_profile_id uuid REFERENCES credential_profiles(id) ON DELETE SET NULL,
	status text NOT NULL,
	target_url text NOT NULL DEFAULT '',
	max_forms integer NOT NULL DEFAULT 10,
	max_tests_per_form integer NOT NULL DEFAULT 1,
	safe_get_only boolean NOT NULL DEFAULT true,
	started_at timestamptz,
	completed_at timestamptz,
	total_forms_detected integer NOT NULL DEFAULT 0,
	total_forms_classified_safe integer NOT NULL DEFAULT 0,
	total_forms_tested integer NOT NULL DEFAULT 0,
	total_forms_skipped integer NOT NULL DEFAULT 0,
	total_findings integer NOT NULL DEFAULT 0,
	error_message text NOT NULL DEFAULT '',
	created_at timestamptz NOT NULL DEFAULT now(),
	updated_at timestamptz NOT NULL DEFAULT now(),
	CONSTRAINT form_test_runs_status_check CHECK (status IN ('queued', 'running', 'completed', 'failed', 'error')),
	CONSTRAINT form_test_runs_max_forms_check CHECK (max_forms >= 1 AND max_forms <= 50),
	CONSTRAINT form_test_runs_max_tests_per_form_check CHECK (max_tests_per_form >= 1 AND max_tests_per_form <= 5)
);

CREATE TABLE IF NOT EXISTS form_test_results (
	id uuid PRIMARY KEY,
	run_id uuid NOT NULL REFERENCES form_test_runs(id) ON DELETE CASCADE,
	project_id uuid NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
	page_url text NOT NULL,
	form_action text NOT NULL DEFAULT '',
	form_method text NOT NULL DEFAULT '',
	classification text NOT NULL DEFAULT 'unknown',
	safety text NOT NULL DEFAULT 'unknown',
	decision text NOT NULL,
	skip_reason text NOT NULL DEFAULT '',
	submitted_url text NOT NULL DEFAULT '',
	final_url text NOT NULL DEFAULT '',
	http_status integer,
	page_title text NOT NULL DEFAULT '',
	test_values_summary jsonb NOT NULL DEFAULT '{}'::jsonb,
	screenshot_evidence_id uuid REFERENCES evidence(id) ON DELETE SET NULL,
	console_error_count integer NOT NULL DEFAULT 0,
	failed_request_count integer NOT NULL DEFAULT 0,
	duration_ms integer,
	finding_id uuid REFERENCES findings(id) ON DELETE SET NULL,
	created_at timestamptz NOT NULL DEFAULT now(),
	CONSTRAINT form_test_results_safety_check CHECK (safety IN ('safe', 'unsafe', 'unsupported', 'unknown')),
	CONSTRAINT form_test_results_decision_check CHECK (decision IN ('tested', 'skipped')),
	CONSTRAINT form_test_results_method_check CHECK (form_method = '' OR form_method = lower(form_method))
);

ALTER TABLE ai_browser_control_runs
	ALTER COLUMN policy_version SET DEFAULT 'v0.22.0-alpha';

ALTER TABLE findings
	ADD COLUMN IF NOT EXISTS form_test_run_id uuid REFERENCES form_test_runs(id) ON DELETE CASCADE;

ALTER TABLE evidence
	ADD COLUMN IF NOT EXISTS form_test_run_id uuid REFERENCES form_test_runs(id) ON DELETE CASCADE;

CREATE INDEX IF NOT EXISTS idx_form_test_runs_project_created_at ON form_test_runs(project_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_form_test_runs_status ON form_test_runs(status);
CREATE INDEX IF NOT EXISTS idx_form_test_results_run_created_at ON form_test_results(run_id, created_at ASC);
CREATE INDEX IF NOT EXISTS idx_form_test_results_project_created_at ON form_test_results(project_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_findings_form_test_run_id ON findings(form_test_run_id);
CREATE INDEX IF NOT EXISTS idx_evidence_form_test_run_id ON evidence(form_test_run_id);
