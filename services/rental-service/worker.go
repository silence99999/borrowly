package main

import (
	"log"
	"time"
)

func (app *application) startReminderWorker() {
	ticker := time.NewTicker(1 * time.Minute)
	go func() {
		for range ticker.C {
			log.Println("rental-service: running reminder check")
			if err := app.sendRentalReminders(); err != nil {
				log.Println("rental-service: reminder worker error:", err)
			}
		}
	}()
}

func (app *application) sendRentalReminders() error {
	reminders, err := app.dbGetRentalsForReminder()
	if err != nil {
		return err
	}
	for _, rem := range reminders {
		// Fetch item title from item-service
		item, err := app.fetchItem(rem.ItemID)
		if err != nil {
			log.Printf("rental-service: could not fetch item %s: %v", rem.ItemID, err)
			continue
		}

		// Fetch renter email from auth-service
		email, err := app.fetchUserEmail(rem.RenterID)
		if err != nil {
			log.Printf("rental-service: could not fetch user %s: %v", rem.RenterID, err)
			continue
		}

		if err := app.sendReminderEmail(email, item.Title, rem.EndAt); err != nil {
			log.Println("rental-service: failed to send reminder:", err)
			continue
		}
		_ = app.dbMarkReminderSent(rem.RentalID)
	}
	return nil
}
