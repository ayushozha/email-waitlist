package handler

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"

	"github.com/ayush10/email-waitlist/internal/middleware"
	"github.com/ayush10/email-waitlist/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SubscribersHandler struct {
	pool *pgxpool.Pool
}

func NewSubscribersHandler(pool *pgxpool.Pool) *SubscribersHandler {
	return &SubscribersHandler{pool: pool}
}

func (h *SubscribersHandler) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	project := middleware.ProjectFromContext(r.Context())
	if project == nil {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	subs, total, err := model.ListSubscribers(r.Context(), h.pool, project.ID, limit, offset)
	if err != nil {
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"subscribers": subs,
		"total":       total,
		"limit":       limit,
		"offset":      offset,
	})
}

func (h *SubscribersHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	project := middleware.ProjectFromContext(r.Context())
	if project == nil {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	email := r.PathValue("email")
	if email == "" {
		http.Error(w, `{"error":"email is required"}`, http.StatusBadRequest)
		return
	}

	if err := model.DeleteSubscriber(r.Context(), h.pool, project.ID, email); err != nil {
		if err.Error() == "subscriber not found" {
			http.Error(w, `{"error":"subscriber not found"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "subscriber removed"})
}

func (h *SubscribersHandler) Export(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	project := middleware.ProjectFromContext(r.Context())
	if project == nil {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	subs, err := model.ExportSubscribersCSV(r.Context(), h.pool, project.ID)
	if err != nil {
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s-subscribers.csv", project.Slug))

	writer := csv.NewWriter(w)
	writer.Write([]string{"email", "metadata", "subscribed_at"})

	for _, s := range subs {
		writer.Write([]string{
			s.Email,
			string(s.Metadata),
			s.SubscribedAt.Format("2006-01-02 15:04:05"),
		})
	}
	writer.Flush()
}
