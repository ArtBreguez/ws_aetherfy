package cronjob

import (
	"wsaetherfy/supabase"
	"log"
	"github.com/robfig/cron/v3"
)

func StartCronJob() {
	c := cron.New(cron.WithSeconds())

	// Rodar uma vez por dia à meia-noite
	_, err := c.AddFunc("@daily", func() {
		log.Println("Running daily cron job to reset API usage...")

		client, err := supabase.InitializeDB()
		if err != nil {
			log.Printf("Error connecting to the database: %v", err)
			return
		}

		err = supabase.ResetApiUsage(client)
		if err != nil {
			log.Printf("Error resetting API usage: %v", err)
			return
		}

		log.Println("API usage has been reset for all users.")
	})

	if err != nil {
		log.Fatalf("Error adding cron job: %v", err)
	}

	c.Start()
	log.Println("Cron job started")

}