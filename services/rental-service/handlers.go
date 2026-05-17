package main

import (
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

func readJSON(r *http.Request, dst any) error {
	return json.NewDecoder(r.Body).Decode(dst)
}

func writeJSON(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, err error) {
	_ = writeJSON(w, status, map[string]string{"error": err.Error()})
}

func calculatePrice(startAt, endAt time.Time, pricePerHour, pricePerDay int, isPlatformItem bool) (total, fee, income int, err error) {
	if !endAt.After(startAt) {
		return 0, 0, 0, errors.New("invalid rental period: end must be after start")
	}
	dur := endAt.Sub(startAt)
	hours := int(math.Ceil(dur.Hours()))
	days := int(math.Ceil(dur.Hours() / 24))

	var candidates []int
	if pricePerHour != 0 {
		candidates = append(candidates, hours*pricePerHour)
	}
	if pricePerDay != 0 {
		candidates = append(candidates, days*pricePerDay)
	}
	if len(candidates) == 0 {
		return 0, 0, 0, errors.New("item has no pricing configured")
	}
	total = candidates[0]
	for _, v := range candidates {
		if v < total {
			total = v
		}
	}
	if isPlatformItem {
		return total, total, 0, nil
	}
	fee = int(math.Round(float64(total) * 0.10))
	income = total - fee
	return total, fee, income, nil
}

func (app *application) createRental(w http.ResponseWriter, r *http.Request) {
	u := r.Context().Value(authUserKey).(AuthUser)
	var in struct {
		ItemID  string `json:"item_id"`
		StartAt string `json:"start_at"`
		EndAt   string `json:"end_at"`
	}
	if err := readJSON(r, &in); err != nil {
		rentalCreations.WithLabelValues("error").Inc()
		writeError(w, http.StatusBadRequest, err)
		return
	}

	item, err := app.fetchItem(in.ItemID)
	if err != nil {
		rentalCreations.WithLabelValues("error").Inc()
		writeError(w, http.StatusNotFound, errors.New("item not found"))
		return
	}

	loc, _ := time.LoadLocation("Asia/Almaty")
	startAt, err := time.ParseInLocation("2006-01-02T15:04", in.StartAt, loc)
	if err != nil {
		rentalCreations.WithLabelValues("error").Inc()
		writeError(w, http.StatusBadRequest, errors.New("invalid start_at format, use 2006-01-02T15:04"))
		return
	}
	endAt, err := time.ParseInLocation("2006-01-02T15:04", in.EndAt, loc)
	if err != nil {
		rentalCreations.WithLabelValues("error").Inc()
		writeError(w, http.StatusBadRequest, errors.New("invalid end_at format, use 2006-01-02T15:04"))
		return
	}

	total, fee, income, err := calculatePrice(startAt, endAt, item.PricePerHour, item.PricePerDay, item.IsPlatformItem)
	if err != nil {
		rentalCreations.WithLabelValues("error").Inc()
		writeError(w, http.StatusBadRequest, err)
		return
	}

	rental := &Rental{
		ItemID:        in.ItemID,
		RenterID:      u.ID,
		PickupPointID: item.PickupPoint.ID,
		StartAt:       startAt,
		EndAt:         endAt,
		TotalPrice:    total,
		PlatformFee:   fee,
		OwnerIncome:   income,
	}
	if err := app.dbCreateRental(rental); err != nil {
		rentalCreations.WithLabelValues("error").Inc()
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	rentalCreations.WithLabelValues("success").Inc()
	_ = writeJSON(w, http.StatusCreated, rental)
}

func (app *application) getMyRentals(w http.ResponseWriter, r *http.Request) {
	u := r.Context().Value(authUserKey).(AuthUser)
	rentals, err := app.dbGetRentalsByUser(u.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	for _, ren := range rentals {
		item, err := app.fetchItem(ren.Item.ID)
		if err != nil {
			continue
		}
		ren.Item.Title = item.Title
		ren.Item.Category = item.Category
		ren.PickupPoint.City = item.PickupPoint.City
		ren.PickupPoint.Address = item.PickupPoint.Address
	}
	_ = writeJSON(w, http.StatusOK, rentals)
}

func (app *application) cancelRental(w http.ResponseWriter, r *http.Request) {
	u := r.Context().Value(authUserKey).(AuthUser)
	id := chi.URLParam(r, "id")

	rental, err := app.dbGetRental(id)
	if err != nil {
		writeError(w, http.StatusNotFound, errors.New("rental not found"))
		return
	}
	if rental.RenterID != u.ID {
		writeError(w, http.StatusForbidden, errors.New("forbidden"))
		return
	}
	if err := app.dbUpdateRentalStatus(u.ID, id, "CANCELLED"); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (app *application) payRental(w http.ResponseWriter, r *http.Request) {
	u := r.Context().Value(authUserKey).(AuthUser)
	id := chi.URLParam(r, "id")

	rental, err := app.dbGetRental(id)
	if err != nil {
		writeError(w, http.StatusNotFound, errors.New("rental not found"))
		return
	}
	if rental.RenterID != u.ID {
		writeError(w, http.StatusForbidden, errors.New("forbidden"))
		return
	}
	// APPROVED is the valid DB status that represents a paid/confirmed rental
	if err := app.dbUpdateRentalStatus(u.ID, id, "APPROVED"); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
