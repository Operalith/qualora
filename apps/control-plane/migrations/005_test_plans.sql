CREATE TABLE IF NOT EXISTS test_plans (
	id uuid PRIMARY KEY,
	project_id uuid NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
	run_id uuid REFERENCES test_runs(id) ON DELETE SET NULL,
	provider_id uuid REFERENCES ai_providers(id) ON DELETE SET NULL,
	model text NOT NULL DEFAULT '',
	status text NOT NULL,
	title text NOT NULL DEFAULT '',
	summary text NOT NULL DEFAULT '',
	plan_json jsonb NOT NULL DEFAULT '{}'::jsonb,
	risk_level text NOT NULL DEFAULT '',
	total_scenarios integer NOT NULL DEFAULT 0,
	error_message text NOT NULL DEFAULT '',
	created_at timestamptz NOT NULL DEFAULT now(),
	updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_test_plans_project_id_created_at ON test_plans(project_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_test_plans_run_id_created_at ON test_plans(run_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_test_plans_status ON test_plans(status);
