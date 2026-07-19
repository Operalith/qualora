package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

func (a *App) createCIRun(w http.ResponseWriter, r *http.Request, projectID string) {
	project, err := a.store.GetProject(r.Context(), projectID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "project_not_found", "project was not found")
			return
		}
		a.logger.Error("get project for CI run failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_project_failed", "project could not be loaded")
		return
	}
	input, err := decodeCIRunRequest(w, r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_ci_run", err.Error())
		return
	}
	input, err = NormalizeCIRunRequest(*project, input)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_ci_run", err.Error())
		return
	}
	created, err := a.store.CreateCIRun(r.Context(), project.ID, map[string]any{"mode": input.Mode})
	if err != nil {
		a.logger.Error("create CI run failed", "error", err)
		writeError(w, http.StatusInternalServerError, "create_ci_run_failed", "CI run could not be created")
		return
	}

	if input.Wait != nil && !*input.Wait {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(input.TimeoutSeconds)*time.Second)
			defer cancel()
			if _, err := a.executeCIRun(ctx, created.ID, *project, input); err != nil {
				a.logger.Error("async CI run failed", "ci_run_id", created.ID, "error", RedactSecrets(err.Error()))
			}
		}()
		writeJSON(w, http.StatusAccepted, ciRunResponse(*created, nil, nil, nil))
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), time.Duration(input.TimeoutSeconds)*time.Second)
	defer cancel()
	response, err := a.executeCIRun(ctx, created.ID, *project, input)
	if response != nil {
		writeJSON(w, http.StatusCreated, response)
		return
	}
	if err != nil {
		a.logger.Error("CI run failed before response", "ci_run_id", created.ID, "error", RedactSecrets(err.Error()))
		writeError(w, http.StatusInternalServerError, "ci_run_failed", "CI run could not be completed")
		return
	}
	writeError(w, http.StatusInternalServerError, "ci_run_failed", "CI run did not produce a response")
}

func (a *App) listCIRuns(w http.ResponseWriter, r *http.Request, projectID string) {
	if _, err := a.store.GetProject(r.Context(), projectID); err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "project_not_found", "project was not found")
			return
		}
		a.logger.Error("get project for CI runs failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_project_failed", "project could not be loaded")
		return
	}
	runs, err := a.store.ListCIRuns(r.Context(), projectID)
	if err != nil {
		a.logger.Error("list CI runs failed", "project_id", projectID, "error", err)
		writeError(w, http.StatusInternalServerError, "list_ci_runs_failed", "CI runs could not be listed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ci_runs": runs})
}

func (a *App) handleCIRunSubroutes(w http.ResponseWriter, r *http.Request) {
	path := stringsTrimPrefix(r.URL.Path, "/api/v1/ci-runs/")
	parts := stringsSplitPath(path)
	if len(parts) != 1 || parts[0] == "" {
		writeError(w, http.StatusNotFound, "not_found", "route not found")
		return
	}
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
		return
	}
	a.getCIRun(w, r, parts[0])
}

func (a *App) getCIRun(w http.ResponseWriter, r *http.Request, ciRunID string) {
	run, err := a.store.GetCIRun(r.Context(), ciRunID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "ci_run_not_found", "CI run was not found")
			return
		}
		a.logger.Error("get CI run failed", "error", err)
		writeError(w, http.StatusInternalServerError, "get_ci_run_failed", "CI run could not be loaded")
		return
	}
	writeJSON(w, http.StatusOK, run)
}

func decodeCIRunRequest(w http.ResponseWriter, r *http.Request) (CIRunRequest, error) {
	var input CIRunRequest
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	if r.Body == nil || r.Body == http.NoBody {
		return input, nil
	}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		if errors.Is(err, io.EOF) {
			return input, nil
		}
		return input, fmt.Errorf("request body must be valid CI run JSON")
	}
	return input, nil
}
