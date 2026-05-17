package main

import (
	"log"
	"time"
)

func (app *application) startRentalReminderWorker() {
	ticker := time.NewTicker(1 * time.Minute)

	go func() {
		for range ticker.C {
			log.Println("tick")
			err := app.sendRentalReminders()
			if err != nil {
				log.Println("reminder worker error:", err)
			}
		}
	}()
}

func (app *application) sendRentalReminders() error {
	rentals, err := app.models.Rentals.GetRentalsForReminder()
	if err != nil {
		return err
	}

	for _, r := range rentals {
		err := app.mailer.SendRentalEndingSoon(r.ItemTitle, r.Email, r.EndAt)
		if err != nil {
			log.Println(err)
			continue
		}

		_ = app.models.Rentals.UpdateReminderSent(r.RentalID)

	}

	return nil
}
