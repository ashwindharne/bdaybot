# BdayBot

## !!! WORK IN PROGRESS !!!

This is a simple ssh app  that reminds you of your friends nad family's birthdays via SMS.
I'm mainly working on it to learn Go and work with
some [Charmbracelet](https://charm.sh) libraries. 

Forgive any Golang atrocities, I'm still learning and will progressively
clean up the implementation.

The app is composed of 2 main components:
1. SSH Server - built with Wish/Bubbletea, reads/writes to a SQLite database.
2. Notification Script - reads the SQlite database and sends SMS reminders via Twilio. Scheduled via cron to run every hour.

It currently only supports running as a standalone Bubbletea app with no SSH server implementation, but
I'm working on integrating it with Wish.