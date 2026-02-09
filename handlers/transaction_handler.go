package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"categories-api/services"
	"categories-api/models"
)

type TransactionHandler struct {
	service *services.TransactionService
}

func NewTransactionHandler(service *services.TransactionService) *TransactionHandler {
	return &TransactionHandler{service: service}
}

// multiple item apa aja, quantity nya
func (h *TransactionHandler) HandleCheckout(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.Checkout(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *TransactionHandler) Checkout(w http.ResponseWriter, r *http.Request) {
	var req models.CheckoutRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	transaction, err := h.service.Checkout(req.Items)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(transaction)
}

func (h *TransactionHandler) HandleReport(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    if r.URL.Path == "/api/report/hari-ini" {
        h.GetTodayReport(w, r)
        return
    }

    // Default
    h.GetRangeReport(w, r)
}

func (h *TransactionHandler) GetTodayReport(w http.ResponseWriter, r *http.Request) {
    today := time.Now().Truncate(24 * time.Hour)
    start := today
    end := today.Add(24 * time.Hour)

    report, err := h.service.GetReport(start, end)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Override
    report.StartDate = today.Format("2006-01-02")
    report.EndDate = today.Format("2006-01-02")

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(report)
}

func (h *TransactionHandler) GetRangeReport(w http.ResponseWriter, r *http.Request) {
    startDateStr := r.URL.Query().Get("start_date")
    endDateStr := r.URL.Query().Get("end_date")

    var start, end time.Time
    var err error

    if startDateStr == "" || endDateStr == "" {
        today := time.Now().Truncate(24 * time.Hour)
        start = today
        end = today.Add(24 * time.Hour)
    } else {
        start, err = time.Parse("2006-01-02", startDateStr)
        if err != nil {
            http.Error(w, "format start_date salah (gunakan YYYY-MM-DD)", http.StatusBadRequest)
            return
        }

        end, err = time.Parse("2006-01-02", endDateStr)
        if err != nil {
            http.Error(w, "format end_date salah (gunakan YYYY-MM-DD)", http.StatusBadRequest)
            return
        }

        end = end.Add(24 * time.Hour)
    }

    if end.Before(start) {
        http.Error(w, "end_date harus setelah atau sama dengan start_date", http.StatusBadRequest)
        return
    }

    report, err := h.service.GetReport(start, end)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    report.StartDate = start.Format("2006-01-02")
    report.EndDate = end.Add(-24 * time.Hour).Format("2006-01-02")

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(report)
}