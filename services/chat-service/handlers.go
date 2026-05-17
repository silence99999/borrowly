package main

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (app *application) resolveEmail(userID string) string {
	resp, err := http.Get(app.authServiceURL + "/internal/users/" + userID)
	if err != nil || resp.StatusCode != http.StatusOK {
		return userID
	}
	defer resp.Body.Close()
	var data struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil || data.Email == "" {
		return userID
	}
	return data.Email
}

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

func (app *application) listConversations(w http.ResponseWriter, r *http.Request) {
	u := r.Context().Value(authUserKey).(AuthUser)
	convs, err := app.dbGetConversations(u.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	for _, c := range convs {
		c.Email = app.resolveEmail(c.UserID)
	}
	_ = writeJSON(w, http.StatusOK, convs)
}

func (app *application) getMessages(w http.ResponseWriter, r *http.Request) {
	u := r.Context().Value(authUserKey).(AuthUser)
	otherUserID := chi.URLParam(r, "userID")

	msgs, err := app.dbGetMessages(u.ID, otherUserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	emailCache := map[string]string{}
	resolve := func(id string) string {
		if e, ok := emailCache[id]; ok {
			return e
		}
		e := app.resolveEmail(id)
		emailCache[id] = e
		return e
	}
	for _, m := range msgs {
		m.SenderEmail = resolve(m.SenderID)
	}
	_ = writeJSON(w, http.StatusOK, msgs)
}

func (app *application) sendMessage(w http.ResponseWriter, r *http.Request) {
	u := r.Context().Value(authUserKey).(AuthUser)
	receiverID := chi.URLParam(r, "userID")

	if receiverID == u.ID {
		writeError(w, http.StatusBadRequest, errors.New("cannot send message to yourself"))
		return
	}

	var in struct {
		Content string `json:"content"`
	}
	if err := readJSON(r, &in); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if len(in.Content) == 0 {
		writeError(w, http.StatusBadRequest, errors.New("content cannot be empty"))
		return
	}

	msg, err := app.dbSendMessage(u.ID, receiverID, in.Content)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	_ = writeJSON(w, http.StatusCreated, msg)
}
