package main

import (
	"database/sql"
	"fmt"
	"os"
)

type reminder struct {
	phoneNumber string
	month       int
	day         int
	year        int
	name        string
}

func main() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "db.sqlite"
	}
	//twilioAccountSid := os.Getenv("TWILIO_ACCOUNT_SID")
	//twilioAuthToken := os.Getenv("TWILIO_AUTH_TOKEN")
	//twilioPhoneNumber := os.Getenv("TWILIO_PHONE_NUMBER")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		panic(err)
	}
	reminderQuery := `
SELECT phone_numbers.phone_number, birthdays.name, birthdays.month, birthdays.day, birthdays.year
FROM birthdays
JOIN phone_numbers ON phone_numbers.id = birthdays.phone_number_id
WHERE 
    phone_numbers.enabled = TRUE
    AND (
        (strftime('%m', 'now') = printf('%02d', birthdays.month) AND 
         cast(strftime('%d', 'now') as integer) <= birthdays.day AND 
         birthdays.day - cast(strftime('%d', 'now') as integer) < phone_numbers.notification_days)
        OR 
        (strftime('%m', 'now') != printf('%02d', birthdays.month) AND 
         (julianday(printf('%04d-%02d-%02d', strftime('%Y', 'now'), birthdays.month, birthdays.day)) - julianday('now')) < phone_numbers.notification_days
        )
    )
    AND cast(strftime('%H', 'now', 'utc') as integer) = phone_numbers.notification_hour_utc;
`
	reminderResults, err := db.Query(reminderQuery)
	if err != nil {
		panic(err)
	}
	defer reminderResults.Close()

	var reminders []reminder
	for reminderResults.Next() {
		reminderResult := reminder{}
		err := reminderResults.Scan(&reminderResult.phoneNumber, &reminderResult.name, &reminderResult.month, &reminderResult.day, &reminderResult.year)
		if err != nil {
			panic(err)
		}
		reminders = append(reminders, reminderResult)
	}

	for _, reminder := range reminders {
		fmt.Printf("Sending reminder to %s for %s's birthday on %d/%d/%d\n", reminder.phoneNumber, reminder.name, reminder.month, reminder.day, reminder.year)
	}
}
