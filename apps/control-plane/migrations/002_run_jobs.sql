CREATE TABLE IF NOT EXISTS run_jobs (
	id uuid PRIMARY KEY,
	run_id uuid NOT NULL REFERENCES test_runs(id) ON DELETE CASCADE,
	kind text NOT NULL,
	status text NOT NULL,
	error_message text NOT NULL DEFAULT '',
	started_at timestamptz,
	completed_at timestamptz,
	created_at timestamptz NOT NULL DEFAULT now(),
	updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_run_jobs_run_id ON run_jobs(run_id);
CREATE INDEX IF NOT EXISTS idx_run_jobs_status ON run_jobs(status);

ALTER TABLE projects
	ALTER COLUMN frontend_url SET DEFAULT '';

CREATE OR REPLACE FUNCTION refresh_test_run_status(run_uuid uuid)
RETURNS text
LANGUAGE plpgsql
AS $$
DECLARE
	total_jobs integer;
	pending_jobs integer;
	running_jobs integer;
	failed_jobs integer;
	next_status text;
	next_error text;
BEGIN
	SELECT
		count(*),
		count(*) FILTER (WHERE status = 'pending'),
		count(*) FILTER (WHERE status = 'running'),
		count(*) FILTER (WHERE status = 'failed')
	INTO total_jobs, pending_jobs, running_jobs, failed_jobs
	FROM run_jobs
	WHERE run_id = run_uuid;

	IF total_jobs = 0 THEN
		next_status := 'failed';
		next_error := 'run has no queued jobs';
	ELSIF pending_jobs > 0 OR running_jobs > 0 THEN
		next_status := 'running';
		next_error := '';
	ELSIF failed_jobs > 0 THEN
		next_status := 'failed';
		SELECT error_message INTO next_error
		FROM run_jobs
		WHERE run_id = run_uuid AND status = 'failed' AND error_message <> ''
		ORDER BY updated_at DESC
		LIMIT 1;
		next_error := COALESCE(next_error, 'one or more run jobs failed');
	ELSE
		next_status := 'completed';
		next_error := '';
	END IF;

	UPDATE test_runs
	SET
		status = next_status,
		error_message = next_error,
		completed_at = CASE
			WHEN next_status IN ('completed', 'failed') THEN COALESCE(completed_at, now())
			ELSE NULL
		END,
		updated_at = now()
	WHERE id = run_uuid;

	RETURN next_status;
END;
$$;
