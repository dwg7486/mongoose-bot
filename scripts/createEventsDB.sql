CREATE TABLE events
(
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    description TEXT NOT NULL,
    location TEXT NOT NULL,
    event_date TEXT NOT NULL,
    event_time TEXT NOT NULL,
    creator TEXT NOT NULL,
    creator_id TEXT NOT NULL
)

CREATE TABLE rsvps
(
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    event_id INTEGER NOT NULL,
    username TEXT NOT NULL,
    user_id TEXT NOT NULL,
    status TEXT NOT NULL,
    FOREIGN KEY (event_id) REFERENCES events (id)
)
