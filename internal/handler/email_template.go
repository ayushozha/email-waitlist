package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/ayush10/email-waitlist/internal/middleware"
	"github.com/ayush10/email-waitlist/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

type EmailTemplateHandler struct {
	pool *pgxpool.Pool
}

func NewEmailTemplateHandler(pool *pgxpool.Pool) *EmailTemplateHandler {
	return &EmailTemplateHandler{pool: pool}
}

func (h *EmailTemplateHandler) Get(w http.ResponseWriter, r *http.Request) {
	project := middleware.ProjectFromContext(r.Context())
	if project == nil {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	tmpl, err := model.GetEmailTemplate(r.Context(), h.pool, project.ID)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"template": nil,
			"message":  "No custom template set. The default confirmation email will be used.",
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"template": tmpl})
}

func (h *EmailTemplateHandler) Upsert(w http.ResponseWriter, r *http.Request) {
	project := middleware.ProjectFromContext(r.Context())
	if project == nil {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var req model.UpsertEmailTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	tmpl, err := model.UpsertEmailTemplate(r.Context(), h.pool, project.ID, req)
	if err != nil {
		log.Printf("upsert email template error [project=%s]: %v", project.Slug, err)
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"message":  "Email template saved.",
		"template": tmpl,
	})
}

func (h *EmailTemplateHandler) Delete(w http.ResponseWriter, r *http.Request) {
	project := middleware.ProjectFromContext(r.Context())
	if project == nil {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	if err := model.DeleteEmailTemplate(r.Context(), h.pool, project.ID); err != nil {
		if err.Error() == "no email template found" {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "no email template found"})
			return
		}
		log.Printf("delete email template error [project=%s]: %v", project.Slug, err)
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "Email template deleted. Default template will be used.",
	})
}
