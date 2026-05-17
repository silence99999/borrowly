package main

import (
	"errors"
	"math"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/silence99999/advanced_final/internal/data"
	"github.com/silence99999/advanced_final/internal/json"
)

func (app *application) CreateRentalHandler(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(authUserKey).(AuthUser)

	var input struct {
		ItemID  string `json:"item_id"`
		StartAt string `json:"start_at"`
		EndAt   string `json:"end_at"`
	}
	err := json.ReadJSON(r, &input)
	if err != nil {
		recordRentalCreationMetric(http.StatusInternalServerError)
		app.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	item, err := app.models.Items.GetByIDRaw(input.ItemID)
	if err != nil {
		recordRentalCreationMetric(http.StatusNotFound)
		app.WriteError(w, http.StatusNotFound, err)
		return
	}

	loc, _ := time.LoadLocation("Asia/Almaty")

	startAt, err := time.ParseInLocation(
		"2006-01-02T15:04",
		input.StartAt,
		loc,
	)

	endAt, err := time.ParseInLocation(
		"2006-01-02T15:04",
		input.EndAt,
		loc,
	)

	totalPrice, platformFee, ownerIncome, err := CalculatePrice(startAt, endAt, item.PricePerHour, item.PricePerDay, item.IsPlatformItem)

	if err != nil {
		recordRentalCreationMetric(http.StatusBadRequest)
		app.WriteError(w, http.StatusBadRequest, err)
		return
	}

	rental := &data.Rental{
		ItemID:        input.ItemID,
		RenterID:      user.ID,
		PickupPointID: item.PickupPointID,
		StartAt:       startAt,
		EndAt:         endAt,
		TotalPrice:    totalPrice,
		PlatformFee:   platformFee,
		OwnerIncome:   ownerIncome,
	}

	err = app.models.Rentals.Create(rental)
	if err != nil {
		recordRentalCreationMetric(http.StatusBadRequest)
		app.WriteError(w, http.StatusBadRequest, err)
		return
	}

	recordRentalCreationMetric(http.StatusCreated)
	err = json.WriteJSON(w, http.StatusCreated, rental)
	if err != nil {
		app.WriteError(w, http.StatusInternalServerError, err)
		return
	}

}

func (app *application) GetMyRentalsHandler(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(authUserKey).(AuthUser)

	rentals, err := app.models.Rentals.GetListByUserID(user.ID)
	if err != nil {
		app.WriteError(w, http.StatusNotFound, err)
		return
	}

	err = json.WriteJSON(w, http.StatusOK, rentals)
	if err != nil {
		app.WriteError(w, http.StatusInternalServerError, err)
		return
	}

}

func (app *application) CancelRentalHandler(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(authUserKey).(AuthUser)

	rentalID := chi.URLParam(r, "id")

	rental, err := app.models.Rentals.GetByID(rentalID)
	if err != nil {
		app.WriteError(w, http.StatusNotFound, err)
		return
	}

	if rental.RenterID != user.ID {
		app.WriteError(w, http.StatusForbidden, errors.New("forbidden"))
		return
	}

	statusCancelled := "CANCELLED"

	err = app.models.Rentals.UpdateStatus(user.ID, rentalID, statusCancelled)
	if err != nil {
		if errors.Is(err, errors.New("Not Found")) {
			app.WriteError(w, http.StatusNotFound, err)
			return
		}
		app.WriteError(w, http.StatusInternalServerError, err)
		return
	}

}

func (app *application) PayRentalHandler(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(authUserKey).(AuthUser)

	rentalID := chi.URLParam(r, "id")

	rental, err := app.models.Rentals.GetByID(rentalID)
	if err != nil {
		app.WriteError(w, http.StatusNotFound, err)
		return
	}

	if rental.RenterID != user.ID {
		app.WriteError(w, http.StatusForbidden, errors.New("forbidden"))
		return
	}

	statusPaid := "PAID"
	err = app.models.Rentals.UpdateStatus(user.ID, rentalID, statusPaid)
	if err != nil {
		if errors.Is(err, errors.New("Not Found")) {
			app.WriteError(w, http.StatusNotFound, err)
			return
		}
		app.WriteError(w, http.StatusInternalServerError, err)
		return
	}

}

func CalculatePrice(startAt, endAt time.Time, pricePerHour, pricePerDay int, isPlatformItem bool) (totalPrice, platformFee, ownerIncome int, err error) {
	if !endAt.After(startAt) {
		return 0, 0, 0, errors.New("Invalid rental period")
	}

	duration := endAt.Sub(startAt)

	hours := int(math.Ceil(duration.Hours()))

	days := int(math.Ceil(duration.Hours() / 24))

	var candidates []int

	if pricePerHour != 0 {
		candidates = append(candidates, hours*pricePerHour)
	}

	if pricePerDay != 0 {
		candidates = append(candidates, days*pricePerDay)
	}

	if len(candidates) == 0 {
		return 0, 0, 0, errors.New("item has no pricing")
	}

	totalPrice = candidates[0]

	for _, value := range candidates {
		if value < totalPrice {
			totalPrice = value
		}
	}

	if isPlatformItem {
		return totalPrice, totalPrice, 0, nil
	}

	platformFee = int(math.Round(float64(totalPrice) * 0.10))
	ownerIncome = totalPrice - platformFee

	return totalPrice, platformFee, ownerIncome, nil
}
