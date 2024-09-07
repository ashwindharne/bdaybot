package main

import (
	"fmt"
	"time"
)

func daysToNextBirthday(bMonth int, bDay int) int {
	now := time.Now()
	nYear, _, _ := now.Date()
	birthdayThisYear := time.Date(nYear, time.Month(bMonth), bDay, 0, 0, 0, 0, time.UTC)
	birthdayNextYear := time.Date(nYear+1, time.Month(bMonth), bDay, 0, 0, 0, 0, time.UTC)
	if birthdayThisYear.After(now) {
		return int(birthdayThisYear.Sub(now).Hours() / 24)
	} else {
		return int(birthdayNextYear.Sub(now).Hours() / 24)
	}
}

func daysTilString(bMonth int, bDay int) string {
	days := daysToNextBirthday(bMonth, bDay)
	if days == 0 {
		return "It's today!"
	} else if days == 1 {
		return "It's tomorrow!"
	} else {
		return fmt.Sprintf("%d days", days)
	}
}
