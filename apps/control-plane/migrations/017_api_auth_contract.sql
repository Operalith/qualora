CREATE TABLE IF NOT EXISTS api_auth_profiles (
	id uuid PRIMARY KEY,
	project_id uuid NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
	name text NOT NULL,
	type text NOT NULL,
	header_name text NOT NULL DEFAULT '',
	query_param_name text NOT NULL DEFAULT '',
	username_encrypted text NOT NULL DEFAULT '',
	password_encrypted text NOT NULL DEFAULT '',
	token_encrypted text NOT NULL DEFAULT '',
	api_key_encrypted text NOT NULL DEFAULT '',
	username_display_hint text NOT NULL DEFAULT '',
	token_display_hint text NOT NULL DEFAULT '',
	api_key_display_hint text NOT NULL DEFAULT '',
	enabled boolean NOT NULL DEFAULT true,
	created_at timestamptz NOT NULL DEFAULT now(),
	updated_at timestamptz NOT NULL DEFAULT now(),
	CONSTRAINT api_auth_profiles_type_check CHECK (type IN ('bearer_token', 'api_key_header', 'api_key_query', 'basic_auth', 'none'))
);

ALTER TABLE test_runs
	ADD COLUMN IF NOT EXISTS api_auth_profile_id uuid REFERENCES api_auth_profiles(id) ON DELETE SET NULL;

ALTER TABLE qa_runs
	ADD COLUMN IF NOT EXISTS api_smoke_run_id uuid REFERENCES test_runs(id) ON DELETE SET NULL,
	ADD COLUMN IF NOT EXISTS api_auth_profile_id uuid REFERENCES api_auth_profiles(id) ON DELETE SET NULL;

ALTER TABLE api_operations
	ADD COLUMN IF NOT EXISTS response_schemas_json jsonb NOT NULL DEFAULT '{}'::jsonb;

ALTER TABLE api_check_results
	ADD COLUMN IF NOT EXISTS api_auth_profile_id uuid REFERENCES api_auth_profiles(id) ON DELETE SET NULL,
	ADD COLUMN IF NOT EXISTS auth_mode text NOT NULL DEFAULT 'none',
	ADD COLUMN IF NOT EXISTS contract_validation_status text NOT NULL DEFAULT 'unknown',
	ADD COLUMN IF NOT EXISTS schema_validation_errors_json jsonb NOT NULL DEFAULT '[]'::jsonb,
	ADD COLUMN IF NOT EXISTS expected_statuses_json jsonb NOT NULL DEFAULT '[]'::jsonb,
	ADD COLUMN IF NOT EXISTS actual_status integer,
	ADD COLUMN IF NOT EXISTS expected_content_types_json jsonb NOT NULL DEFAULT '[]'::jsonb,
	ADD COLUMN IF NOT EXISTS actual_content_type text NOT NULL DEFAULT '',
	ADD COLUMN IF NOT EXISTS response_time_ms integer,
	ADD COLUMN IF NOT EXISTS unauthenticated_status integer;

CREATE INDEX IF NOT EXISTS idx_api_auth_profiles_project_id_created_at ON api_auth_profiles(project_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_api_auth_profiles_enabled ON api_auth_profiles(enabled);
CREATE INDEX IF NOT EXISTS idx_test_runs_api_auth_profile_id ON test_runs(api_auth_profile_id);
CREATE INDEX IF NOT EXISTS idx_qa_runs_api_smoke_run_id ON qa_runs(api_smoke_run_id);
CREATE INDEX IF NOT EXISTS idx_qa_runs_api_auth_profile_id ON qa_runs(api_auth_profile_id);
CREATE INDEX IF NOT EXISTS idx_api_check_results_api_auth_profile_id ON api_check_results(api_auth_profile_id);
CREATE INDEX IF NOT EXISTS idx_api_check_results_contract_validation_status ON api_check_results(contract_validation_status);
