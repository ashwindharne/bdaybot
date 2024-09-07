CREATE TABLE IF NOT EXISTS phone_numbers
(
    id                INTEGER PRIMARY KEY AUTOINCREMENT,
    phone_number      TEXT    UNIQUE NOT NULL,
    verified          BOOLEAN NOT NULL DEFAULT FALSE,
    notification_days INTEGER NOT NULL DEFAULT 14,
    display_timezone         TEXT    NOT NULL DEFAULT 'EST',
    notification_hour_utc INTEGER NOT NULL DEFAULT 7,
    enabled           BOOLEAN NOT NULL DEFAULT TRUE,
    created_at        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS birthdays
(
    id                         INTEGER PRIMARY KEY AUTOINCREMENT,
    name                       TEXT    NOT NULL,
    month                      INTEGER NOT NULL,
    day                        INTEGER NOT NULL,
    year                       INTEGER NOT NULL,
    phone_number_id            INTEGER NOT NULL,
    created_at                 DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at                 DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (phone_number_id) REFERENCES phone_numbers (id)
);

CREATE INDEX IF NOT EXISTS phone_numbers_phone_number ON phone_numbers (phone_number);
CREATE INDEX IF NOT EXISTS birthdays_phone_number_id ON birthdays (phone_number_id);

