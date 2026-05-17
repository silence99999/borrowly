package main

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/silence99999/advanced_final/internal/data"
	"github.com/silence99999/advanced_final/internal/json"
)

func (app *application) CreateItemHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		PickupPointID string `json:"pickup_point_id"`
		Title         string `json:"title"`
		Description   string `json:"description"`
		Category      string `json:"category"`
		PricePerHour  int    `json:"price_per_hour"`
		PricePerDay   int    `json:"price_per_day"`
	}

	err := json.ReadJSON(r, &input)
	if err != nil {
		app.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	if _, err := uuid.Parse(input.PickupPointID); err != nil {
		app.WriteError(w, http.StatusBadRequest, errors.New("pickup_point_id must be a valid UUID"))
		return
	}

	user := r.Context().Value(authUserKey).(AuthUser)

	item := &data.Item{
		OwnerID:       user.ID,
		PickupPointID: input.PickupPointID,
		Title:         input.Title,
		Description:   input.Description,
		Category:      input.Category,
		PricePerHour:  input.PricePerHour,
		PricePerDay:   input.PricePerDay,
	}

	if user.Role == "ADMIN" {
		item.IsPlatformItem = true
	}

	err = app.models.Items.Create(item)
	if err != nil {
		app.WriteError(w, http.StatusBadRequest, err)
		return
	}

	err = json.WriteJSON(w, http.StatusCreated, item)
	if err != nil {
		app.WriteError(w, http.StatusInternalServerError, err)
		return
	}

}

func (app *application) GetItemsHandler(w http.ResponseWriter, r *http.Request) {
	items, err := app.models.Items.GetAll()
	if err != nil {
		recordItemBrowseMetric("list", http.StatusInternalServerError)
		app.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	recordItemBrowseMetric("list", http.StatusOK)
	err = json.WriteJSON(w, http.StatusOK, items)
	if err != nil {
		app.WriteError(w, http.StatusInternalServerError, err)
		return
	}

}

func (app *application) GetItemByIDHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	item, err := app.models.Items.GetByID(id)
	if err != nil {
		recordItemBrowseMetric("details", http.StatusNotFound)
		app.WriteError(w, http.StatusNotFound, err)
		return
	}

	recordItemBrowseMetric("details", http.StatusOK)
	err = json.WriteJSON(w, http.StatusOK, item)
	if err != nil {
		app.WriteError(w, http.StatusInternalServerError, err)
		return
	}

}

func (app *application) UpdateItemHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var input struct {
		PickupPointID *string `json:"pickup_point_id"`
		Title         *string `json:"title"`
		Description   *string `json:"description"`
		Category      *string `json:"category"`
		PricePerHour  *int    `json:"price_per_hour"`
		PricePerDay   *int    `json:"price_per_day"`
	}

	err := json.ReadJSON(r, &input)
	if err != nil {
		app.WriteError(w, http.StatusBadRequest, err)
		return
	}

	authUser := r.Context().Value(authUserKey).(AuthUser)

	item, err := app.models.Items.GetByIDRaw(id)
	if err != nil {
		app.WriteError(w, http.StatusNotFound, errors.New("item not found"))
		return
	}

	if item.OwnerID != authUser.ID && authUser.Role != "ADMIN" {
		app.WriteError(w, http.StatusForbidden, errors.New("forbidden"))
		return
	}

	if input.PricePerHour != nil && *input.PricePerHour <= 0 {
		app.WriteError(w, http.StatusBadRequest, errors.New("price_per_hour must be > 0"))
		return
	}

	if input.PricePerDay != nil && *input.PricePerDay <= 0 {
		app.WriteError(w, http.StatusBadRequest, errors.New("price_per_day must be > 0"))
		return
	}

	if input.PickupPointID != nil {
		if _, err := uuid.Parse(*input.PickupPointID); err != nil {
			app.WriteError(w, http.StatusBadRequest, errors.New("pickup_point_id must be a valid UUID"))
			return
		}
	}

	if input.PickupPointID == nil &&
		input.Title == nil &&
		input.Description == nil &&
		input.Category == nil &&
		input.PricePerHour == nil &&
		input.PricePerDay == nil {
		app.WriteError(w, http.StatusBadRequest, errors.New("no fields to update"))
		return
	}

	itemUpdate := data.ItemUpdate{
		PickupPointID: input.PickupPointID,
		Title:         input.Title,
		Description:   input.Description,
		Category:      input.Category,
		PricePerHour:  input.PricePerHour,
		PricePerDay:   input.PricePerDay,
	}

	err = app.models.Items.UpdateByID(id, itemUpdate)
	if err != nil {
		app.WriteError(w, http.StatusInternalServerError, err)
		return
	}

}

func (app *application) DeleteItemHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	authUser := r.Context().Value(authUserKey).(AuthUser)

	item, err := app.models.Items.GetByIDRaw(id)
	if err != nil {
		app.WriteError(w, http.StatusNotFound, errors.New("item not found"))
		return
	}

	if item.OwnerID != authUser.ID && authUser.Role != "ADMIN" {
		app.WriteError(w, http.StatusForbidden, errors.New("forbidden"))
		return
	}

	err = app.models.Items.DeleteByID(id)
	if err != nil {
		app.WriteError(w, http.StatusInternalServerError, err)
		return
	}
}
func (app *application) GetMyItemsHandler(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(authUserKey).(AuthUser)

	items, err := app.models.Items.GetByOwnerID(user.ID)
	if err != nil {
		app.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	err = json.WriteJSON(w, http.StatusOK, items)
	if err != nil {
		app.WriteError(w, http.StatusInternalServerError, err)
		return
	}
}
