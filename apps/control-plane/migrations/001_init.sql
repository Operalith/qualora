CREATE TABLE IF NOT EXISTS projects (
	id uuid PRIMARY KEY,
	name text NOT NULL,
	frontend_url text NOT NULL,
	api_base_url text NOT NULL DEFAULT '',
	openapi_url text NOT NULL DEFAULT '',
	allowed_hosts jsonb NOT NULL DEFAULT '[]'::jsonb,
	security_mode text NOT NULL DEFAULT 'passive',
	destructive_actions boolean NOT NULL DEFAULT false,
	allow_private_targets boolean NOT NULL DEFAULT false,
	created_at timestamptz NOT NULL DEFAULT now(),
	updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS test_runs (
	id uuid PRIMARY KEY,
	project_id uuid NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
	status text NOT NULL,
	error_message text NOT NULL DEFAULT '',
	page_title text NOT NULL DEFAULT '',
	started_at timestamptz,
	completed_at timestamptz,
	created_at timestamptz NOT NULL DEFAULT now(),
	updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS findings (
	id uuid PRIMARY KEY,
	run_id uuid NOT NULL REFERENCES test_runs(id) ON DELETE CASCADE,
	title text NOT NULL,
	severity text NOT NULL,
	category text NOT NULL,
	confidence text NOT NULL,
	description text NOT NULL,
	recommendation text NOT NULL,
	evidence_ids jsonb NOT NULL DEFAULT '[]'::jsonb,
	created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS evidence (
	id uuid PRIMARY KEY,
	run_id uuid NOT NULL REFERENCES test_runs(id) ON DELETE CASCADE,
	type text NOT NULL,
	uri text NOT NULL,
	metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
	created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_projects_created_at ON projects(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_test_runs_project_id ON test_runs(project_id);
CREATE INDEX IF NOT EXISTS idx_test_runs_status ON test_runs(status);
CREATE INDEX IF NOT EXISTS idx_findings_run_id ON findings(run_id);
CREATE INDEX IF NOT EXISTS idx_evidence_run_id ON evidence(run_id);
