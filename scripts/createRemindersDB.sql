CREATE TABLE reminders
(
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id TEXT NOT NULL,
  remind_datetime TEXT NOT NULL,
  message TEXT NOT NULL
)