CREATE TABLE IF NOT EXISTS report_baselines (
	id uuid PRIMARY KEY,
	project_id uuid NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
	name text NOT NULL,
	description text,
	report_type text NOT NULL,
	report_id text NOT NULL,
	source_run_id text,
	fingerprint_set_json jsonb NOT NULL DEFAULT '[]'::jsonb,
	severity_counts_json jsonb NOT NULL DEFAULT '{}'::jsonb,
	grouped_findings_count integer NOT NULL DEFAULT 0,
	raw_findings_count integer NOT NULL DEFAULT 0,
	created_by_user_id uuid REFERENCES local_users(id) ON DELETE SET NULL,
	is_default boolean NOT NULL DEFAULT false,
	created_at timestamptz NOT NULL DEFAULT now(),
	updated_at timestamptz NOT NULL DEFAULT now(),
	CONSTRAINT report_baselines_report_type_check CHECK (report_type IN (
		'safe_qa',
		'quality_check',
		'discovery',
		'safe_explorer',
		'api_smoke',
		'browser_smoke',
		'authorization'
	)),
	CONSTRAINT report_baselines_name_length_check CHECK (char_length(name) BETWEEN 1 AND 160)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_report_baselines_one_default_per_type
	ON report_baselines(project_id, report_type)
	WHERE is_default;

CREATE INDEX IF NOT EXISTS idx_report_baselines_project_type_created_at
	ON report_baselines(project_id, report_type, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_report_baselines_report_lookup
	ON report_baselines(project_id, report_type, report_id);
