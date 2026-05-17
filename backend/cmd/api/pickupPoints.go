package main

import (
	"net/http"

	"github.com/silence99999/advanced_final/internal/data"
	"github.com/silence99999/advanced_final/internal/json"
)

func (app *application) GetPickupPoints(w http.ResponseWriter, r *http.Request) {
	pickupPoints, err := app.models.PickupPoints.GetAllActive()
	if err != nil {
		app.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	err = json.WriteJSON(w, http.StatusOK, pickupPoints)
	if err != nil {
		app.WriteError(w, http.StatusInternalServerError, err)
		return
	}
}

func (app *application) CreatePickupPoint(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Address  string `json:"address"`
		City     string `json:"city"`
		IsActive bool   `json:"is_active"`
	}

	err := json.ReadJSON(r, &input)
	if err != nil {
		app.WriteError(w, http.StatusBadRequest, err)
		return
	}

	pickupPoint := &data.PickupPoint{
		Address:  input.Address,
		City:     input.City,
		IsActive: input.IsActive,
	}

	err = app.models.PickupPoints.Create(pickupPoint)
	if err != nil {
		app.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	err = json.WriteJSON(w, http.StatusCreated, pickupPoint)
	if err != nil {
		app.WriteError(w, http.StatusInternalServerError, err)
		return
	}
}
