package handler

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"

	"github.com/ayush10/email-waitlist/internal/email"
	"github.com/ayush10/email-waitlist/internal/middleware"
	"github.com/ayush10/email-waitlist/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SubscribeHandler struct {
	pool         *pgxpool.Pool
	emailService *email.Service
}

func NewSubscribeHandler(pool *pgxpool.Pool, emailService *email.Service) *SubscribeHandler {
	return &SubscribeHandler{pool: pool, emailService: emailService}
}

func (h *SubscribeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	project := middleware.ProjectFromContext(r.Context())
	if project == nil {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	// Limit request body to 10KB to prevent abuse
	r.Body = http.MaxBytesReader(w, r.Body, 10*1024)

	var req model.SubscribeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		if err == io.EOF {
			http.Error(w, `{"error":"empty request body"}`, http.StatusBadRequest)
			return
		}
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if err := req.Validate(); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	// Limit metadata size to 4KB
	if len(req.Metadata) > 4096 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "metadata too large (max 4KB)"})
		return
	}

	sub, err := model.AddSubscriber(r.Context(), h.pool, project.ID, req)
	if err != nil {
		switch {
		case errors.Is(err, model.ErrAlreadySubscribed):
			writeJSON(w, http.StatusConflict, map[string]string{
				"error":   "already subscribed",
				"message": "This email is already on the waitlist.",
			})
		case errors.Is(err, model.ErrReferralCodeTaken):
			writeJSON(w, http.StatusConflict, map[string]string{
				"error":   "referral code already taken",
				"message": "This referral code is already in use for this project.",
			})
		default:
			log.Printf("subscribe error [project=%s]: %v", project.Slug, err)
			http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		}
		return
	}

	// Send confirmation email asynchronously
	if h.emailService != nil {
		go h.emailService.SendConfirmation(project, sub)
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"message":    "Successfully joined the waitlist!",
		"subscriber": sub,
	})
}
