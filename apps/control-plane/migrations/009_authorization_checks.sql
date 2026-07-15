ALTER TABLE credential_profiles
	ADD COLUMN IF NOT EXISTS role_name text NOT NULL DEFAULT '',
	ADD COLUMN IF NOT EXISTS role_description text NOT NULL DEFAULT '',
	ADD COLUMN IF NOT EXISTS subject_label text NOT NULL DEFAULT '';

CREATE TABLE IF NOT EXISTS authorization_checks (
	id uuid PRIMARY KEY,
	project_id uuid NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
	name text NOT NULL,
	description text NOT NULL DEFAULT '',
	type text NOT NULL,
	resource_label text NOT NULL DEFAULT '',
	owner_credential_profile_id uuid REFERENCES credential_profiles(id) ON DELETE SET NULL,
	actor_credential_profile_id uuid NOT NULL REFERENCES credential_profiles(id) ON DELETE CASCADE,
	expected_outcome text NOT NULL,
	target_url text NOT NULL DEFAULT '',
	api_spec_id uuid REFERENCES api_specs(id) ON DELETE SET NULL,
	api_operation_id uuid REFERENCES api_operations(id) ON DELETE SET NULL,
	method text NOT NULL DEFAULT '',
	path text NOT NULL DEFAULT '',
	expected_statuses_json jsonb NOT NULL DEFAULT '[]'::jsonb,
	success_text_contains text NOT NULL DEFAULT '',
	denied_statuses_json jsonb NOT NULL DEFAULT '[]'::jsonb,
	denied_text_contains text NOT NULL DEFAULT '',
	enabled boolean NOT NULL DEFAULT true,
	created_at timestamptz NOT NULL DEFAULT now(),
	updated_at timestamptz NOT NULL DEFAULT now(),
	CONSTRAINT authorization_checks_type_check CHECK (type IN ('browser_url', 'api_get')),
	CONSTRAINT authorization_checks_expected_outcome_check CHECK (expected_outcome IN ('allowed', 'denied'))
);

CREATE TABLE IF NOT EXISTS authorization_check_runs (
	id uuid PRIMARY KEY,
	project_id uuid NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
	status text NOT NULL,
	check_ids_json jsonb NOT NULL DEFAULT '[]'::jsonb,
	max_checks integer NOT NULL DEFAULT 10,
	total_checks integer NOT NULL DEFAULT 0,
	passed_checks integer NOT NULL DEFAULT 0,
	failed_checks integer NOT NULL DEFAULT 0,
	skipped_checks integer NOT NULL DEFAULT 0,
	error_message text NOT NULL DEFAULT '',
	started_at timestamptz,
	completed_at timestamptz,
	created_at timestamptz NOT NULL DEFAULT now(),
	updated_at timestamptz NOT NULL DEFAULT now(),
	CONSTRAINT authorization_check_runs_status_check CHECK (status IN ('queued', 'running', 'completed', 'failed', 'error')),
	CONSTRAINT authorization_check_runs_max_checks_check CHECK (max_checks >= 1 AND max_checks <= 50)
);

CREATE TABLE IF NOT EXISTS authorization_check_results (
	id uuid PRIMARY KEY,
	run_id uuid NOT NULL REFERENCES authorization_check_runs(id) ON DELETE CASCADE,
	check_id uuid NOT NULL REFERENCES authorization_checks(id) ON DELETE CASCADE,
	status text NOT NULL,
	expected_outcome text NOT NULL,
	actual_outcome text NOT NULL,
	actor_credential_profile_id uuid NOT NULL REFERENCES credential_profiles(id) ON DELETE CASCADE,
	actor_role_name text NOT NULL DEFAULT '',
	target_url text NOT NULL DEFAULT '',
	final_url text NOT NULL DEFAULT '',
	http_status integer,
	page_title text NOT NULL DEFAULT '',
	duration_ms integer,
	evidence_id uuid REFERENCES evidence(id) ON DELETE SET NULL,
	finding_id uuid REFERENCES findings(id) ON DELETE SET NULL,
	skip_reason text NOT NULL DEFAULT '',
	error_message text NOT NULL DEFAULT '',
	created_at timestamptz NOT NULL DEFAULT now(),
	CONSTRAINT authorization_check_results_status_check CHECK (status IN ('passed', 'failed', 'skipped', 'error')),
	CONSTRAINT authorization_check_results_expected_outcome_check CHECK (expected_outcome IN ('allowed', 'denied')),
	CONSTRAINT authorization_check_results_actual_outcome_check CHECK (actual_outcome IN ('allowed', 'denied', 'unknown'))
);

ALTER TABLE findings
	ADD COLUMN IF NOT EXISTS authorization_check_run_id uuid REFERENCES authorization_check_runs(id) ON DELETE CASCADE;

ALTER TABLE evidence
	ADD COLUMN IF NOT EXISTS authorization_check_run_id uuid REFERENCES authorization_check_runs(id) ON DELETE CASCADE;

CREATE INDEX IF NOT EXISTS idx_credential_profiles_project_role_name ON credential_profiles(project_id, role_name);
CREATE INDEX IF NOT EXISTS idx_authorization_checks_project_created_at ON authorization_checks(project_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_authorization_checks_project_enabled ON authorization_checks(project_id, enabled);
CREATE INDEX IF NOT EXISTS idx_authorization_checks_actor_profile ON authorization_checks(actor_credential_profile_id);
CREATE INDEX IF NOT EXISTS idx_authorization_check_runs_project_created_at ON authorization_check_runs(project_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_authorization_check_runs_status ON authorization_check_runs(status);
CREATE INDEX IF NOT EXISTS idx_authorization_check_results_run_id ON authorization_check_results(run_id);
CREATE INDEX IF NOT EXISTS idx_authorization_check_results_check_id ON authorization_check_results(check_id);
CREATE INDEX IF NOT EXISTS idx_authorization_check_results_evidence_id ON authorization_check_results(evidence_id);
CREATE INDEX IF NOT EXISTS idx_authorization_check_results_finding_id ON authorization_check_results(finding_id);
CREATE INDEX IF NOT EXISTS idx_findings_authorization_check_run_id ON findings(authorization_check_run_id);
CREATE INDEX IF NOT EXISTS idx_evidence_authorization_check_run_id ON evidence(authorization_check_run_id);
