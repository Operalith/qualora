CREATE TABLE IF NOT EXISTS discovery_runs (
	id uuid PRIMARY KEY,
	project_id uuid NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
	credential_profile_id uuid REFERENCES credential_profiles(id) ON DELETE SET NULL,
	status text NOT NULL,
	start_url text NOT NULL,
	max_pages integer NOT NULL DEFAULT 20,
	max_depth integer NOT NULL DEFAULT 2,
	same_origin_only boolean NOT NULL DEFAULT true,
	started_at timestamptz,
	completed_at timestamptz,
	total_pages integer NOT NULL DEFAULT 0,
	total_links integer NOT NULL DEFAULT 0,
	total_forms integer NOT NULL DEFAULT 0,
	total_console_errors integer NOT NULL DEFAULT 0,
	total_failed_requests integer NOT NULL DEFAULT 0,
	total_findings integer NOT NULL DEFAULT 0,
	error_message text NOT NULL DEFAULT '',
	created_at timestamptz NOT NULL DEFAULT now(),
	updated_at timestamptz NOT NULL DEFAULT now(),
	CONSTRAINT discovery_runs_status_check CHECK (status IN ('queued', 'running', 'completed', 'failed', 'error')),
	CONSTRAINT discovery_runs_max_pages_check CHECK (max_pages >= 1 AND max_pages <= 100),
	CONSTRAINT discovery_runs_max_depth_check CHECK (max_depth >= 0 AND max_depth <= 5)
);

ALTER TABLE findings
	ADD COLUMN IF NOT EXISTS discovery_run_id uuid REFERENCES discovery_runs(id) ON DELETE CASCADE;

ALTER TABLE evidence
	ADD COLUMN IF NOT EXISTS discovery_run_id uuid REFERENCES discovery_runs(id) ON DELETE CASCADE;

CREATE TABLE IF NOT EXISTS discovered_pages (
	id uuid PRIMARY KEY,
	discovery_run_id uuid NOT NULL REFERENCES discovery_runs(id) ON DELETE CASCADE,
	project_id uuid NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
	url text NOT NULL,
	normalized_url text NOT NULL,
	path text NOT NULL DEFAULT '',
	title text NOT NULL DEFAULT '',
	http_status integer,
	content_type text NOT NULL DEFAULT '',
	body_text_length integer,
	load_duration_ms integer,
	depth integer NOT NULL DEFAULT 0,
	screenshot_evidence_id uuid REFERENCES evidence(id) ON DELETE SET NULL,
	console_error_count integer NOT NULL DEFAULT 0,
	failed_request_count integer NOT NULL DEFAULT 0,
	discovered_at timestamptz NOT NULL DEFAULT now(),
	created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS discovered_links (
	id uuid PRIMARY KEY,
	discovery_run_id uuid NOT NULL REFERENCES discovery_runs(id) ON DELETE CASCADE,
	source_page_id uuid NOT NULL REFERENCES discovered_pages(id) ON DELETE CASCADE,
	href text NOT NULL,
	normalized_url text NOT NULL DEFAULT '',
	link_text text NOT NULL DEFAULT '',
	same_origin boolean NOT NULL DEFAULT false,
	skipped boolean NOT NULL DEFAULT false,
	skip_reason text NOT NULL DEFAULT '',
	created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS discovered_forms (
	id uuid PRIMARY KEY,
	discovery_run_id uuid NOT NULL REFERENCES discovery_runs(id) ON DELETE CASCADE,
	page_id uuid NOT NULL REFERENCES discovered_pages(id) ON DELETE CASCADE,
	form_name text NOT NULL DEFAULT '',
	form_action text NOT NULL DEFAULT '',
	form_method text NOT NULL DEFAULT '',
	field_count integer NOT NULL DEFAULT 0,
	password_field_count integer NOT NULL DEFAULT 0,
	submit_button_count integer NOT NULL DEFAULT 0,
	classification text NOT NULL DEFAULT '',
	skipped_reason text NOT NULL DEFAULT '',
	created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS discovered_form_fields (
	id uuid PRIMARY KEY,
	form_id uuid NOT NULL REFERENCES discovered_forms(id) ON DELETE CASCADE,
	field_name text NOT NULL DEFAULT '',
	field_type text NOT NULL DEFAULT '',
	placeholder text NOT NULL DEFAULT '',
	label text NOT NULL DEFAULT '',
	required boolean NOT NULL DEFAULT false,
	created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_discovery_runs_project_created_at ON discovery_runs(project_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_discovery_runs_status ON discovery_runs(status);
CREATE INDEX IF NOT EXISTS idx_discovered_pages_run_depth ON discovered_pages(discovery_run_id, depth, created_at);
CREATE INDEX IF NOT EXISTS idx_discovered_pages_normalized_url ON discovered_pages(discovery_run_id, normalized_url);
CREATE INDEX IF NOT EXISTS idx_discovered_links_run ON discovered_links(discovery_run_id);
CREATE INDEX IF NOT EXISTS idx_discovered_links_source_page ON discovered_links(source_page_id);
CREATE INDEX IF NOT EXISTS idx_discovered_forms_run ON discovered_forms(discovery_run_id);
CREATE INDEX IF NOT EXISTS idx_discovered_form_fields_form ON discovered_form_fields(form_id);
CREATE INDEX IF NOT EXISTS idx_findings_discovery_run_id ON findings(discovery_run_id);
CREATE INDEX IF NOT EXISTS idx_evidence_discovery_run_id ON evidence(discovery_run_id);
