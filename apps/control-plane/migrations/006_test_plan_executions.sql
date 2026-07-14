CREATE TABLE IF NOT EXISTS test_plan_executions (
	id uuid PRIMARY KEY,
	test_plan_id uuid NOT NULL REFERENCES test_plans(id) ON DELETE CASCADE,
	project_id uuid NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
	source_run_id uuid REFERENCES test_runs(id) ON DELETE SET NULL,
	status text NOT NULL,
	total_scenarios integer NOT NULL DEFAULT 0,
	passed_scenarios integer NOT NULL DEFAULT 0,
	failed_scenarios integer NOT NULL DEFAULT 0,
	skipped_scenarios integer NOT NULL DEFAULT 0,
	total_steps integer NOT NULL DEFAULT 0,
	passed_steps integer NOT NULL DEFAULT 0,
	failed_steps integer NOT NULL DEFAULT 0,
	skipped_steps integer NOT NULL DEFAULT 0,
	error_message text NOT NULL DEFAULT '',
	started_at timestamptz,
	completed_at timestamptz,
	created_at timestamptz NOT NULL DEFAULT now(),
	updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS test_plan_execution_scenarios (
	id uuid PRIMARY KEY,
	execution_id uuid NOT NULL REFERENCES test_plan_executions(id) ON DELETE CASCADE,
	scenario_id_from_plan text NOT NULL,
	name text NOT NULL,
	type text NOT NULL DEFAULT '',
	priority text NOT NULL DEFAULT '',
	status text NOT NULL,
	skip_reason text NOT NULL DEFAULT '',
	started_at timestamptz,
	completed_at timestamptz,
	created_at timestamptz NOT NULL DEFAULT now(),
	updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS test_plan_execution_steps (
	id uuid PRIMARY KEY,
	execution_id uuid NOT NULL REFERENCES test_plan_executions(id) ON DELETE CASCADE,
	scenario_execution_id uuid NOT NULL REFERENCES test_plan_execution_scenarios(id) ON DELETE CASCADE,
	step_order integer NOT NULL,
	original_action text NOT NULL DEFAULT '',
	mapped_action text NOT NULL DEFAULT '',
	target text NOT NULL DEFAULT '',
	expected_result text NOT NULL DEFAULT '',
	status text NOT NULL,
	skip_reason text NOT NULL DEFAULT '',
	actual_result text NOT NULL DEFAULT '',
	error_message text NOT NULL DEFAULT '',
	duration_ms integer,
	evidence_id uuid,
	created_at timestamptz NOT NULL DEFAULT now(),
	updated_at timestamptz NOT NULL DEFAULT now()
);

ALTER TABLE findings ALTER COLUMN run_id DROP NOT NULL;
ALTER TABLE findings
	ADD COLUMN IF NOT EXISTS test_plan_execution_id uuid REFERENCES test_plan_executions(id) ON DELETE CASCADE,
	ADD COLUMN IF NOT EXISTS scenario_execution_id uuid REFERENCES test_plan_execution_scenarios(id) ON DELETE SET NULL,
	ADD COLUMN IF NOT EXISTS step_execution_id uuid REFERENCES test_plan_execution_steps(id) ON DELETE SET NULL;

ALTER TABLE evidence ALTER COLUMN run_id DROP NOT NULL;
ALTER TABLE evidence
	ADD COLUMN IF NOT EXISTS test_plan_execution_id uuid REFERENCES test_plan_executions(id) ON DELETE CASCADE;

CREATE INDEX IF NOT EXISTS idx_test_plan_executions_test_plan_id_created_at ON test_plan_executions(test_plan_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_test_plan_executions_project_id_created_at ON test_plan_executions(project_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_test_plan_executions_status ON test_plan_executions(status);
CREATE INDEX IF NOT EXISTS idx_test_plan_execution_scenarios_execution_id ON test_plan_execution_scenarios(execution_id);
CREATE INDEX IF NOT EXISTS idx_test_plan_execution_steps_execution_id ON test_plan_execution_steps(execution_id);
CREATE INDEX IF NOT EXISTS idx_test_plan_execution_steps_scenario_execution_id ON test_plan_execution_steps(scenario_execution_id);
CREATE INDEX IF NOT EXISTS idx_findings_test_plan_execution_id ON findings(test_plan_execution_id);
CREATE INDEX IF NOT EXISTS idx_evidence_test_plan_execution_id ON evidence(test_plan_execution_id);

CREATE OR REPLACE FUNCTION refresh_test_plan_execution_status(execution_uuid uuid)
RETURNS text AS $$
DECLARE
	total_scenario_count integer;
	passed_scenario_count integer;
	failed_scenario_count integer;
	skipped_scenario_count integer;
	total_step_count integer;
	passed_step_count integer;
	failed_step_count integer;
	skipped_step_count integer;
	active_step_count integer;
	running_step_count integer;
	current_status text;
	next_status text;
BEGIN
	SELECT status
	INTO current_status
	FROM test_plan_executions
	WHERE id = execution_uuid;

	SELECT
		count(*),
		count(*) FILTER (WHERE status = 'passed'),
		count(*) FILTER (WHERE status IN ('failed', 'error')),
		count(*) FILTER (WHERE status = 'skipped')
	INTO total_scenario_count, passed_scenario_count, failed_scenario_count, skipped_scenario_count
	FROM test_plan_execution_scenarios
	WHERE execution_id = execution_uuid;

	SELECT
		count(*),
		count(*) FILTER (WHERE status = 'passed'),
		count(*) FILTER (WHERE status IN ('failed', 'error')),
		count(*) FILTER (WHERE status = 'skipped'),
		count(*) FILTER (WHERE status IN ('queued', 'pending', 'running')),
		count(*) FILTER (WHERE status = 'running')
	INTO total_step_count, passed_step_count, failed_step_count, skipped_step_count, active_step_count, running_step_count
	FROM test_plan_execution_steps
	WHERE execution_id = execution_uuid;

	IF total_step_count = 0 THEN
		next_status := 'completed';
	ELSIF active_step_count > 0 THEN
		IF current_status = 'queued' AND running_step_count = 0 THEN
			next_status := 'queued';
		ELSE
			next_status := 'running';
		END IF;
	ELSIF failed_step_count > 0 THEN
		next_status := 'failed';
	ELSE
		next_status := 'completed';
	END IF;

	UPDATE test_plan_executions
	SET
		status = next_status,
		total_scenarios = total_scenario_count,
		passed_scenarios = passed_scenario_count,
		failed_scenarios = failed_scenario_count,
		skipped_scenarios = skipped_scenario_count,
		total_steps = total_step_count,
		passed_steps = passed_step_count,
		failed_steps = failed_step_count,
		skipped_steps = skipped_step_count,
		completed_at = CASE
			WHEN next_status IN ('completed', 'failed') THEN COALESCE(completed_at, now())
			ELSE completed_at
		END,
		updated_at = now()
	WHERE id = execution_uuid;

	RETURN next_status;
END;
$$ LANGUAGE plpgsql;
