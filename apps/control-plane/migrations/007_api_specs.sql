CREATE TABLE IF NOT EXISTS api_specs (
	id uuid PRIMARY KEY,
	project_id uuid NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
	name text NOT NULL,
	source_type text NOT NULL,
	source_url text,
	raw_spec text,
	parsed_title text NOT NULL DEFAULT '',
	parsed_version text NOT NULL DEFAULT '',
	server_url text NOT NULL DEFAULT '',
	operation_count integer NOT NULL DEFAULT 0,
	safe_operation_count integer NOT NULL DEFAULT 0,
	skipped_operation_count integer NOT NULL DEFAULT 0,
	status text NOT NULL,
	error_message text NOT NULL DEFAULT '',
	created_at timestamptz NOT NULL DEFAULT now(),
	updated_at timestamptz NOT NULL DEFAULT now(),
	CONSTRAINT api_specs_source_type_check CHECK (source_type IN ('url', 'inline', 'demo')),
	CONSTRAINT api_specs_status_check CHECK (status IN ('pending', 'parsed', 'invalid', 'error'))
);

CREATE TABLE IF NOT EXISTS api_operations (
	id uuid PRIMARY KEY,
	api_spec_id uuid NOT NULL REFERENCES api_specs(id) ON DELETE CASCADE,
	project_id uuid NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
	method text NOT NULL,
	path text NOT NULL,
	resolved_path text NOT NULL DEFAULT '',
	query_string text NOT NULL DEFAULT '',
	operation_id text NOT NULL DEFAULT '',
	summary text NOT NULL DEFAULT '',
	description text NOT NULL DEFAULT '',
	tags_json jsonb NOT NULL DEFAULT '[]'::jsonb,
	expected_statuses_json jsonb NOT NULL DEFAULT '[]'::jsonb,
	expected_content_types_json jsonb NOT NULL DEFAULT '[]'::jsonb,
	requires_authentication boolean,
	safe_to_execute boolean NOT NULL DEFAULT false,
	skip_reason text NOT NULL DEFAULT '',
	created_at timestamptz NOT NULL DEFAULT now(),
	updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS api_check_results (
	id uuid PRIMARY KEY,
	run_id uuid NOT NULL REFERENCES test_runs(id) ON DELETE CASCADE,
	api_spec_id uuid NOT NULL REFERENCES api_specs(id) ON DELETE CASCADE,
	operation_id uuid REFERENCES api_operations(id) ON DELETE SET NULL,
	method text NOT NULL,
	path text NOT NULL,
	resolved_url text NOT NULL DEFAULT '',
	status text NOT NULL,
	http_status integer,
	duration_ms integer,
	response_content_type text NOT NULL DEFAULT '',
	response_size_bytes integer,
	error_message text NOT NULL DEFAULT '',
	skipped_reason text NOT NULL DEFAULT '',
	created_at timestamptz NOT NULL DEFAULT now()
);

ALTER TABLE test_runs
	ADD COLUMN IF NOT EXISTS run_type text NOT NULL DEFAULT 'full',
	ADD COLUMN IF NOT EXISTS api_spec_id uuid REFERENCES api_specs(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_api_specs_project_id_created_at ON api_specs(project_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_api_specs_status ON api_specs(status);
CREATE INDEX IF NOT EXISTS idx_api_operations_api_spec_id ON api_operations(api_spec_id);
CREATE INDEX IF NOT EXISTS idx_api_operations_project_id ON api_operations(project_id);
CREATE INDEX IF NOT EXISTS idx_api_operations_safe_to_execute ON api_operations(safe_to_execute);
CREATE INDEX IF NOT EXISTS idx_api_check_results_run_id ON api_check_results(run_id);
CREATE INDEX IF NOT EXISTS idx_api_check_results_api_spec_id ON api_check_results(api_spec_id);
CREATE INDEX IF NOT EXISTS idx_test_runs_run_type ON test_runs(run_type);
CREATE INDEX IF NOT EXISTS idx_test_runs_api_spec_id ON test_runs(api_spec_id);
