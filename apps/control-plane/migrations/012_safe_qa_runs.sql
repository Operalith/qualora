ALTER TABLE test_plans
	ADD COLUMN IF NOT EXISTS discovery_run_id uuid REFERENCES discovery_runs(id) ON DELETE SET NULL,
	ADD COLUMN IF NOT EXISTS source_type text NOT NULL DEFAULT '',
	ADD COLUMN IF NOT EXISTS execution_coverage_json jsonb NOT NULL DEFAULT '{}'::jsonb;

CREATE INDEX IF NOT EXISTS idx_test_plans_discovery_run_id_created_at ON test_plans(discovery_run_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_test_plans_source_type ON test_plans(source_type);

CREATE TABLE IF NOT EXISTS qa_runs (
	id uuid PRIMARY KEY,
	project_id uuid NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
	status text NOT NULL,
	mode text NOT NULL DEFAULT 'safe',
	discovery_run_id uuid REFERENCES discovery_runs(id) ON DELETE SET NULL,
	test_plan_id uuid REFERENCES test_plans(id) ON DELETE SET NULL,
	test_plan_execution_id uuid REFERENCES test_plan_executions(id) ON DELETE SET NULL,
	credential_profile_id uuid REFERENCES credential_profiles(id) ON DELETE SET NULL,
	started_at timestamptz,
	completed_at timestamptz,
	error_message text NOT NULL DEFAULT '',
	summary_json jsonb NOT NULL DEFAULT '{}'::jsonb,
	created_at timestamptz NOT NULL DEFAULT now(),
	updated_at timestamptz NOT NULL DEFAULT now(),
	CONSTRAINT qa_runs_mode_check CHECK (mode IN ('safe')),
	CONSTRAINT qa_runs_status_check CHECK (status IN (
		'queued',
		'running_discovery',
		'generating_plan',
		'previewing_execution',
		'executing_plan',
		'completed',
		'failed',
		'error'
	))
);

CREATE INDEX IF NOT EXISTS idx_qa_runs_project_created_at ON qa_runs(project_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_qa_runs_status ON qa_runs(status);
CREATE INDEX IF NOT EXISTS idx_qa_runs_discovery_run_id ON qa_runs(discovery_run_id);
CREATE INDEX IF NOT EXISTS idx_qa_runs_test_plan_id ON qa_runs(test_plan_id);
CREATE INDEX IF NOT EXISTS idx_qa_runs_test_plan_execution_id ON qa_runs(test_plan_execution_id);
