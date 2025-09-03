package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/tmozzze/order_checker/internal/models"
	"github.com/tmozzze/order_checker/internal/service"
)

type OrderHandler struct {
	service *service.OrderService
}

func NewOrderHandler(svc *service.OrderService) *OrderHandler {
	return &OrderHandler{service: svc}
}

func (h *OrderHandler) RegisterRoutes(r chi.Router) {
	r.Post("/orders", h.SaveOrder)
	r.Get("/orders/{id}", h.GetOrder)
}

func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "missing order id", http.StatusBadRequest)
		return
	}

	order, err := h.service.GetOrder(r.Context(), id)
	if err != nil {
		http.Error(w, "order not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)

}

func (h *OrderHandler) SaveOrder(w http.ResponseWriter, r *http.Request) {
	var order models.Order

	// Decode json
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := order.Validate(); err != nil {
		http.Error(w, "invalid order: "+err.Error(), http.StatusBadRequest)
		log.Printf("Validation failed: %v", err)
		return
	}

	// Service layer
	ctx := r.Context()
	if err := h.service.SaveOrder(ctx, &order); err != nil {
		http.Error(w, "failed to save order", http.StatusInternalServerError)
		log.Printf("Postgres error: %v", err)
		return
	}

	// Success
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	log.Printf("Order %s saved", order.OrderUID)
}
