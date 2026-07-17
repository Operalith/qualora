CREATE TABLE IF NOT EXISTS safe_explorer_runs (
	id uuid PRIMARY KEY,
	project_id uuid NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
	credential_profile_id uuid REFERENCES credential_profiles(id) ON DELETE SET NULL,
	status text NOT NULL,
	start_url text NOT NULL,
	max_steps integer NOT NULL DEFAULT 10,
	max_depth integer NOT NULL DEFAULT 2,
	same_origin_only boolean NOT NULL DEFAULT true,
	allow_get_forms boolean NOT NULL DEFAULT false,
	started_at timestamptz,
	completed_at timestamptz,
	total_steps integer NOT NULL DEFAULT 0,
	total_pages_observed integer NOT NULL DEFAULT 0,
	total_actions_detected integer NOT NULL DEFAULT 0,
	total_actions_executed integer NOT NULL DEFAULT 0,
	total_actions_skipped integer NOT NULL DEFAULT 0,
	total_findings integer NOT NULL DEFAULT 0,
	error_message text NOT NULL DEFAULT '',
	created_at timestamptz NOT NULL DEFAULT now(),
	updated_at timestamptz NOT NULL DEFAULT now(),
	CONSTRAINT safe_explorer_runs_status_check CHECK (status IN ('queued', 'running', 'completed', 'failed', 'error')),
	CONSTRAINT safe_explorer_runs_max_steps_check CHECK (max_steps >= 1 AND max_steps <= 50),
	CONSTRAINT safe_explorer_runs_max_depth_check CHECK (max_depth >= 0 AND max_depth <= 5)
);

ALTER TABLE findings
	ADD COLUMN IF NOT EXISTS safe_explorer_run_id uuid REFERENCES safe_explorer_runs(id) ON DELETE CASCADE;

ALTER TABLE evidence
	ADD COLUMN IF NOT EXISTS safe_explorer_run_id uuid REFERENCES safe_explorer_runs(id) ON DELETE CASCADE;

CREATE TABLE IF NOT EXISTS safe_explorer_steps (
	id uuid PRIMARY KEY,
	run_id uuid NOT NULL REFERENCES safe_explorer_runs(id) ON DELETE CASCADE,
	project_id uuid NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
	step_index integer NOT NULL,
	page_url text NOT NULL,
	normalized_url text NOT NULL,
	page_title text NOT NULL DEFAULT '',
	depth integer NOT NULL DEFAULT 0,
	action_id uuid,
	action_type text NOT NULL DEFAULT '',
	action_label text NOT NULL DEFAULT '',
	action_selector_hint text NOT NULL DEFAULT '',
	action_target_url text NOT NULL DEFAULT '',
	action_safety text NOT NULL DEFAULT 'unknown',
	action_decision text NOT NULL DEFAULT 'observed',
	skip_reason text NOT NULL DEFAULT '',
	result_status text NOT NULL DEFAULT 'ok',
	http_status integer,
	final_url text NOT NULL DEFAULT '',
	screenshot_evidence_id uuid REFERENCES evidence(id) ON DELETE SET NULL,
	console_error_count integer NOT NULL DEFAULT 0,
	failed_request_count integer NOT NULL DEFAULT 0,
	duration_ms integer,
	created_at timestamptz NOT NULL DEFAULT now(),
	CONSTRAINT safe_explorer_steps_safety_check CHECK (action_safety IN ('safe', 'unsafe', 'unsupported', 'unknown')),
	CONSTRAINT safe_explorer_steps_decision_check CHECK (action_decision IN ('executed', 'skipped', 'observed')),
	CONSTRAINT safe_explorer_steps_result_status_check CHECK (result_status IN ('ok', 'failed', 'skipped', 'error')),
	CONSTRAINT safe_explorer_steps_step_index_check CHECK (step_index >= 0),
	CONSTRAINT safe_explorer_steps_depth_check CHECK (depth >= 0)
);

CREATE TABLE IF NOT EXISTS safe_explorer_actions (
	id uuid PRIMARY KEY,
	run_id uuid NOT NULL REFERENCES safe_explorer_runs(id) ON DELETE CASCADE,
	step_id uuid NOT NULL REFERENCES safe_explorer_steps(id) ON DELETE CASCADE,
	source_url text NOT NULL,
	action_type text NOT NULL,
	label text NOT NULL DEFAULT '',
	text text NOT NULL DEFAULT '',
	selector_hint text NOT NULL DEFAULT '',
	href text NOT NULL DEFAULT '',
	target_url text NOT NULL DEFAULT '',
	method text NOT NULL DEFAULT '',
	same_origin boolean NOT NULL DEFAULT false,
	safety text NOT NULL DEFAULT 'unknown',
	decision text NOT NULL DEFAULT 'skip',
	skip_reason text NOT NULL DEFAULT '',
	created_at timestamptz NOT NULL DEFAULT now(),
	CONSTRAINT safe_explorer_actions_type_check CHECK (action_type IN ('link_navigation', 'button', 'form_get', 'form_post', 'input', 'unknown')),
	CONSTRAINT safe_explorer_actions_safety_check CHECK (safety IN ('safe', 'unsafe', 'unsupported', 'unknown')),
	CONSTRAINT safe_explorer_actions_decision_check CHECK (decision IN ('execute', 'skip'))
);

CREATE INDEX IF NOT EXISTS idx_safe_explorer_runs_project_created_at ON safe_explorer_runs(project_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_safe_explorer_runs_status ON safe_explorer_runs(status);
CREATE INDEX IF NOT EXISTS idx_safe_explorer_steps_run_index ON safe_explorer_steps(run_id, step_index);
CREATE INDEX IF NOT EXISTS idx_safe_explorer_steps_normalized_url ON safe_explorer_steps(run_id, normalized_url);
CREATE INDEX IF NOT EXISTS idx_safe_explorer_actions_run ON safe_explorer_actions(run_id, created_at);
CREATE INDEX IF NOT EXISTS idx_safe_explorer_actions_step ON safe_explorer_actions(step_id);
CREATE INDEX IF NOT EXISTS idx_findings_safe_explorer_run_id ON findings(safe_explorer_run_id);
CREATE INDEX IF NOT EXISTS idx_evidence_safe_explorer_run_id ON evidence(safe_explorer_run_id);
