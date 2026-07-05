package handler

import (
	"encoding/csv"
	"fmt"
	"log"
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
	project := middleware.ProjectFromContext(r.Context())
	if project == nil {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	subs, total, err := model.ListSubscribers(r.Context(), h.pool, project.ID, limit, offset)
	if err != nil {
		log.Printf("list subscribers error [project=%s]: %v", project.Slug, err)
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
		log.Printf("delete subscriber error [project=%s]: %v", project.Slug, err)
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "subscriber removed"})
}

func (h *SubscribersHandler) Export(w http.ResponseWriter, r *http.Request) {
	project := middleware.ProjectFromContext(r.Context())
	if project == nil {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	subs, err := model.ExportSubscribersCSV(r.Context(), h.pool, project.ID)
	if err != nil {
		log.Printf("export subscribers error [project=%s]: %v", project.Slug, err)
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s-subscribers.csv", project.Slug))

	writer := csv.NewWriter(w)
	writer.Write([]string{"email", "metadata", "subscribed_at"})

	for _, s := range subs {
		writer.Write([]string{
			csvCell(s.Email),
			csvCell(string(s.Metadata)),
			s.SubscribedAt.Format("2006-01-02 15:04:05"),
		})
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		// Headers are already sent; the download is just truncated.
		log.Printf("export csv write error [project=%s]: %v", project.Slug, err)
	}
}

// csvCell neutralizes spreadsheet formula injection: subscriber-controlled
// values starting with =, +, -, @, tab, or CR would otherwise execute as
// formulas when the export is opened in Excel or Sheets.
func csvCell(s string) string {
	if s == "" {
		return s
	}
	switch s[0] {
	case '=', '+', '-', '@', '\t', '\r':
		return "'" + s
	}
	return s
}
