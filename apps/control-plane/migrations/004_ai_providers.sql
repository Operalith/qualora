CREATE TABLE IF NOT EXISTS ai_providers (
	id uuid PRIMARY KEY,
	name text NOT NULL,
	preset text NOT NULL DEFAULT 'custom',
	type text NOT NULL,
	base_url text NOT NULL,
	model text NOT NULL,
	api_key_encrypted text NOT NULL DEFAULT '',
	extra_headers_encrypted text NOT NULL DEFAULT '',
	temperature double precision NOT NULL DEFAULT 0.2,
	max_output_tokens integer NOT NULL DEFAULT 1200,
	timeout_seconds integer NOT NULL DEFAULT 30,
	send_screenshots boolean NOT NULL DEFAULT false,
	send_html boolean NOT NULL DEFAULT false,
	send_network_bodies boolean NOT NULL DEFAULT false,
	redaction_enabled boolean NOT NULL DEFAULT true,
	is_default boolean NOT NULL DEFAULT false,
	created_at timestamptz NOT NULL DEFAULT now(),
	updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_ai_providers_created_at ON ai_providers(created_at DESC);
CREATE UNIQUE INDEX IF NOT EXISTS idx_ai_providers_one_default ON ai_providers(is_default) WHERE is_default;

CREATE TABLE IF NOT EXISTS ai_analyses (
	id uuid PRIMARY KEY,
	run_id uuid NOT NULL REFERENCES test_runs(id) ON DELETE CASCADE,
	provider_id uuid REFERENCES ai_providers(id) ON DELETE SET NULL,
	model text NOT NULL,
	status text NOT NULL,
	executive_summary text NOT NULL DEFAULT '',
	technical_summary text NOT NULL DEFAULT '',
	risk_level text NOT NULL DEFAULT '',
	analysis_json jsonb NOT NULL DEFAULT '{}'::jsonb,
	prompt_tokens integer NOT NULL DEFAULT 0,
	completion_tokens integer NOT NULL DEFAULT 0,
	total_tokens integer NOT NULL DEFAULT 0,
	error_message text NOT NULL DEFAULT '',
	created_at timestamptz NOT NULL DEFAULT now(),
	updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_ai_analyses_run_id_created_at ON ai_analyses(run_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_ai_analyses_status ON ai_analyses(status);
