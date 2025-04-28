package handler

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"frappuccino/internal/service"
)

type StatsHandler struct {
	StatsSvc service.StatsService
}

func NewStatsHandler(svc service.StatsService) *StatsHandler {
	return &StatsHandler{
		StatsSvc: svc,
	}
}

func (t *StatsHandler) GetPopularItem(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	data, err := t.StatsSvc.GetPopularItem(ctx)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			Respond(w, http.StatusRequestTimeout, "Request timeout")
			return
		}
		Err := FromError(err)
		Respond(w, Err.Status, Err.Message)
		return
	}

	Respond(w, http.StatusOK, data)
}

func (t *StatsHandler) GetTotalSum(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	data, err := t.StatsSvc.GetTotalSum(ctx)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			Respond(w, http.StatusRequestTimeout, "Request timeout")
			return
		}
		Err := FromError(err)
		Respond(w, Err.Status, Err.Message)
		return
	}
	Respond(w, http.StatusOK, data)
}

func (t *StatsHandler) GetSearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if query == "" {
		Respond(w, http.StatusBadRequest, "Query parameter 'q' is required")
		return
	}

	filter := r.URL.Query().Get("filter")

	var minPrice, maxPrice *float64
	if min := r.URL.Query().Get("minPrice"); min != "" {
		minVal, err := strconv.ParseFloat(min, 64)
		if err == nil {
			minPrice = &minVal
		}
	}
	if max := r.URL.Query().Get("maxPrice"); max != "" {
		maxVal, err := strconv.ParseFloat(max, 64)
		if err == nil {
			maxPrice = &maxVal
		}
	}

	results, err := t.StatsSvc.GetSearch(ctx, query, filter, minPrice, maxPrice)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			Respond(w, http.StatusRequestTimeout, "Request timeout")
			return
		}
		Err := FromError(err)
		Respond(w, Err.Status, Err.Message)
		return
	}

	Respond(w, http.StatusOK, results)
}

func (t *StatsHandler) GetItemByPeriod(w http.ResponseWriter, r *http.Request) {
	period := r.URL.Query().Get("period")
	month := r.URL.Query().Get("month")
	yearStr := r.URL.Query().Get("year")

	var year int
	if yearStr != "" {
		y, err := strconv.Atoi(yearStr)
		if err != nil {
			Respond(w, http.StatusBadRequest, "Invalid year format")
			return
		}
		year = y
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	list, err := t.StatsSvc.GetItemByPeriod(ctx, period, month, year)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			Respond(w, http.StatusRequestTimeout, "Request timeout")
			return
		}
		Err := FromError(err)
		Respond(w, Err.Status, Err.Message)
		return
	}

	Respond(w, http.StatusOK, list)
}
