package main

import (
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/silence99999/advanced_final/internal/data"
	"github.com/silence99999/advanced_final/internal/json"
)

func (app *application) CreateItemImageHandler(w http.ResponseWriter, r *http.Request) {
	itemID := chi.URLParam(r, "id")

	user := r.Context().Value(authUserKey).(AuthUser)

	item, err := app.models.Items.GetByIDRaw(itemID)
	if err != nil {
		app.WriteError(w, http.StatusBadRequest, err)
		return
	}

	if item.OwnerID != user.ID {
		app.WriteError(w, http.StatusBadRequest, errors.New("you cant add image to not your item"))
		return
	}

	err = r.ParseMultipartForm(10 << 20)
	if err != nil {
		app.WriteError(w, http.StatusBadRequest, err)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		app.WriteError(w, http.StatusBadRequest, err)
		return
	}
	defer file.Close()

	uploadDir := "uploads/item_images"

	err = os.MkdirAll(uploadDir, os.ModePerm)
	if err != nil {
		app.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	safeName := filepath.Base(header.Filename)

	fullPath := filepath.Join(uploadDir, safeName)

	dst, err := os.Create(fullPath)
	if err != nil {
		app.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		app.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	itemImage := &data.ItemImage{
		ItemID:   itemID,
		FileName: safeName,
	}

	err = app.models.ItemImages.Create(itemImage)
	if err != nil {
		app.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	err = json.WriteJSON(w, http.StatusOK, map[string]string{
		"message": "uploaded successfully",
	})
	if err != nil {
		app.WriteError(w, http.StatusInternalServerError, err)
		return
	}
}
