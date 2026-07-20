CREATE TABLE IF NOT EXISTS ai_browser_control_runs (
	id uuid PRIMARY KEY,
	project_id uuid NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
	provider_id uuid NOT NULL REFERENCES ai_providers(id) ON DELETE RESTRICT,
	credential_profile_id uuid REFERENCES credential_profiles(id) ON DELETE SET NULL,
	status text NOT NULL,
	start_url text NOT NULL,
	goal text NOT NULL DEFAULT '',
	max_steps integer NOT NULL DEFAULT 8,
	max_depth integer NOT NULL DEFAULT 2,
	same_origin_only boolean NOT NULL DEFAULT true,
	policy_version text NOT NULL DEFAULT 'v0.21.0-alpha',
	execution_mode text NOT NULL DEFAULT 'policy_gated',
	started_at timestamptz,
	completed_at timestamptz,
	total_steps integer NOT NULL DEFAULT 0,
	total_ai_suggestions integer NOT NULL DEFAULT 0,
	total_actions_approved integer NOT NULL DEFAULT 0,
	total_actions_executed integer NOT NULL DEFAULT 0,
	total_actions_skipped integer NOT NULL DEFAULT 0,
	total_policy_blocks integer NOT NULL DEFAULT 0,
	total_findings integer NOT NULL DEFAULT 0,
	error_message text NOT NULL DEFAULT '',
	created_at timestamptz NOT NULL DEFAULT now(),
	updated_at timestamptz NOT NULL DEFAULT now(),
	CONSTRAINT ai_browser_control_runs_status_check CHECK (status IN ('queued', 'running', 'completed', 'failed', 'error')),
	CONSTRAINT ai_browser_control_runs_execution_mode_check CHECK (execution_mode IN ('policy_gated')),
	CONSTRAINT ai_browser_control_runs_max_steps_check CHECK (max_steps >= 1 AND max_steps <= 30),
	CONSTRAINT ai_browser_control_runs_max_depth_check CHECK (max_depth >= 0 AND max_depth <= 5)
);

ALTER TABLE ai_browser_control_runs
	ADD COLUMN IF NOT EXISTS policy_version text NOT NULL DEFAULT 'v0.21.0-alpha',
	ADD COLUMN IF NOT EXISTS execution_mode text NOT NULL DEFAULT 'policy_gated';

ALTER TABLE findings
	ADD COLUMN IF NOT EXISTS ai_browser_control_run_id uuid REFERENCES ai_browser_control_runs(id) ON DELETE CASCADE;

ALTER TABLE evidence
	ADD COLUMN IF NOT EXISTS ai_browser_control_run_id uuid REFERENCES ai_browser_control_runs(id) ON DELETE CASCADE;

CREATE TABLE IF NOT EXISTS ai_browser_control_steps (
	id uuid PRIMARY KEY,
	run_id uuid NOT NULL REFERENCES ai_browser_control_runs(id) ON DELETE CASCADE,
	project_id uuid NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
	step_index integer NOT NULL,
	page_url text NOT NULL,
	normalized_url text NOT NULL,
	page_title text NOT NULL DEFAULT '',
	depth integer NOT NULL DEFAULT 0,
	sanitized_observation_json jsonb NOT NULL DEFAULT '{}'::jsonb,
	ai_suggestion_json jsonb,
	action_type text NOT NULL DEFAULT '',
	action_label text NOT NULL DEFAULT '',
	action_target_url text NOT NULL DEFAULT '',
	action_selector_hint text NOT NULL DEFAULT '',
	policy_decision text NOT NULL DEFAULT 'skipped',
	policy_reason text NOT NULL DEFAULT '',
	execution_status text NOT NULL DEFAULT 'skipped',
	final_url text NOT NULL DEFAULT '',
	http_status integer,
	screenshot_evidence_id uuid REFERENCES evidence(id) ON DELETE SET NULL,
	console_error_count integer NOT NULL DEFAULT 0,
	failed_request_count integer NOT NULL DEFAULT 0,
	duration_ms integer,
	created_at timestamptz NOT NULL DEFAULT now(),
	CONSTRAINT ai_browser_control_steps_policy_check CHECK (policy_decision IN ('approved', 'blocked', 'unsupported', 'invalid', 'skipped')),
	CONSTRAINT ai_browser_control_steps_execution_check CHECK (execution_status IN ('executed', 'skipped', 'failed', 'error')),
	CONSTRAINT ai_browser_control_steps_step_index_check CHECK (step_index >= 0),
	CONSTRAINT ai_browser_control_steps_depth_check CHECK (depth >= 0)
);

CREATE INDEX IF NOT EXISTS idx_ai_browser_control_runs_project_created_at ON ai_browser_control_runs(project_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_ai_browser_control_runs_status ON ai_browser_control_runs(status);
CREATE INDEX IF NOT EXISTS idx_ai_browser_control_steps_run_index ON ai_browser_control_steps(run_id, step_index);
CREATE INDEX IF NOT EXISTS idx_ai_browser_control_steps_normalized_url ON ai_browser_control_steps(run_id, normalized_url);
CREATE INDEX IF NOT EXISTS idx_findings_ai_browser_control_run_id ON findings(ai_browser_control_run_id);
CREATE INDEX IF NOT EXISTS idx_evidence_ai_browser_control_run_id ON evidence(ai_browser_control_run_id);
