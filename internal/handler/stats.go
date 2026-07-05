package handler

import (
	"log"
	"net/http"

	"github.com/ayush10/email-waitlist/internal/middleware"
	"github.com/ayush10/email-waitlist/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

type StatsHandler struct {
	pool *pgxpool.Pool
}

func NewStatsHandler(pool *pgxpool.Pool) *StatsHandler {
	return &StatsHandler{pool: pool}
}

func (h *StatsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	project := middleware.ProjectFromContext(r.Context())
	if project == nil {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	stats, err := model.GetStats(r.Context(), h.pool, project.ID)
	if err != nil {
		log.Printf("stats error [project=%s]: %v", project.Slug, err)
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, stats)
}
