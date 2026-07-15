CREATE TABLE IF NOT EXISTS credential_profiles (
	id uuid PRIMARY KEY,
	project_id uuid NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
	name text NOT NULL,
	type text NOT NULL,
	username_encrypted text NOT NULL DEFAULT '',
	password_encrypted text NOT NULL DEFAULT '',
	username_display_hint text NOT NULL DEFAULT '',
	login_url text NOT NULL,
	username_selector text NOT NULL,
	password_selector text NOT NULL,
	submit_selector text NOT NULL,
	success_url_contains text NOT NULL DEFAULT '',
	success_text_contains text NOT NULL DEFAULT '',
	failure_text_contains text NOT NULL DEFAULT '',
	post_login_wait_ms integer NOT NULL DEFAULT 0,
	is_default boolean NOT NULL DEFAULT false,
	created_at timestamptz NOT NULL DEFAULT now(),
	updated_at timestamptz NOT NULL DEFAULT now(),
	CONSTRAINT credential_profiles_type_check CHECK (type IN ('username_password')),
	CONSTRAINT credential_profiles_post_login_wait_check CHECK (post_login_wait_ms >= 0 AND post_login_wait_ms <= 30000)
);

ALTER TABLE test_runs
	ADD COLUMN IF NOT EXISTS credential_profile_id uuid REFERENCES credential_profiles(id) ON DELETE SET NULL,
	ADD COLUMN IF NOT EXISTS target_path text NOT NULL DEFAULT '',
	ADD COLUMN IF NOT EXISTS capture_screenshot boolean NOT NULL DEFAULT true,
	ADD COLUMN IF NOT EXISTS max_duration_seconds integer NOT NULL DEFAULT 30;

CREATE INDEX IF NOT EXISTS idx_credential_profiles_project_id_created_at ON credential_profiles(project_id, created_at DESC);
CREATE UNIQUE INDEX IF NOT EXISTS idx_credential_profiles_project_default
	ON credential_profiles(project_id)
	WHERE is_default;
CREATE INDEX IF NOT EXISTS idx_test_runs_credential_profile_id ON test_runs(credential_profile_id);
