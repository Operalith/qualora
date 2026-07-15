CREATE TABLE IF NOT EXISTS local_users (
	id uuid PRIMARY KEY,
	email text NOT NULL UNIQUE,
	display_name text NOT NULL,
	password_hash text NOT NULL,
	role text NOT NULL DEFAULT 'admin',
	is_active boolean NOT NULL DEFAULT true,
	last_login_at timestamptz,
	created_at timestamptz NOT NULL DEFAULT now(),
	updated_at timestamptz NOT NULL DEFAULT now(),
	CONSTRAINT local_users_role_check CHECK (role IN ('admin'))
);

CREATE TABLE IF NOT EXISTS user_sessions (
	id uuid PRIMARY KEY,
	user_id uuid NOT NULL REFERENCES local_users(id) ON DELETE CASCADE,
	token_hash text NOT NULL UNIQUE,
	csrf_token_hash text,
	user_agent text NOT NULL DEFAULT '',
	ip_address text NOT NULL DEFAULT '',
	expires_at timestamptz NOT NULL,
	created_at timestamptz NOT NULL DEFAULT now(),
	last_seen_at timestamptz NOT NULL DEFAULT now(),
	revoked_at timestamptz
);

CREATE INDEX IF NOT EXISTS idx_local_users_email ON local_users(lower(email));
CREATE INDEX IF NOT EXISTS idx_user_sessions_user_id ON user_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_user_sessions_token_hash ON user_sessions(token_hash);
CREATE INDEX IF NOT EXISTS idx_user_sessions_active ON user_sessions(expires_at) WHERE revoked_at IS NULL;
