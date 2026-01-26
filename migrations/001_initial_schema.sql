CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS events (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    timezone TEXT NOT NULL,
    venue TEXT, -- Added for context if needed
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS reservations (
    id TEXT PRIMARY KEY, -- Using TEXT for UUID string
    user_id TEXT NOT NULL, -- references users(id)
    event_id TEXT NOT NULL, -- references events(id)
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE NOT NULL,
    ticket_count INT NOT NULL DEFAULT 1,
    status TEXT NOT NULL,
    version INT NOT NULL DEFAULT 1,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_reservations_event_time ON reservations (event_id, start_time);

-- Data Migration: Migrate legacy provider IDs to event IDs
UPDATE reservations SET event_id = 'event-1' WHERE event_id = 'provider-1';
UPDATE reservations SET event_id = 'event-2' WHERE event_id = 'provider-2';
UPDATE reservations SET event_id = 'event-3' WHERE event_id = 'provider-3';
