package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/ayush10/email-waitlist/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProjectsHandler struct {
	pool *pgxpool.Pool
}

func NewProjectsHandler(pool *pgxpool.Pool) *ProjectsHandler {
	return &ProjectsHandler{pool: pool}
}

func (h *ProjectsHandler) Create(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 16*1024)

	var req model.CreateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if err := req.Validate(); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	project, err := model.CreateProject(r.Context(), h.pool, req)
	if err != nil {
		if errors.Is(err, model.ErrSlugTaken) {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "project slug already exists"})
			return
		}
		log.Printf("create project error: %v", err)
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"message": "Project created. Save the secret api_key — it won't be shown again. The public_key is safe to embed in your frontend.",
		"project": project,
	})
}

func (h *ProjectsHandler) List(w http.ResponseWriter, r *http.Request) {
	projects, err := model.ListProjects(r.Context(), h.pool)
	if err != nil {
		log.Printf("list projects error: %v", err)
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"projects": projects})
}
