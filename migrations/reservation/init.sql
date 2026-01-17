-- ======================================
-- Reservation Domain Schema
-- ======================================
-- Schema for the Reservation bounded context.
-- This script runs automatically on first PostgreSQL startup.

-- ======================================
-- Reservations Table
-- ======================================
-- Stores the Reservation aggregate root.
-- Status transitions: pending -> confirmed -> active -> completed (or cancelled)
CREATE TABLE IF NOT EXISTS reservations (
    id VARCHAR(255) PRIMARY KEY,
    guest_id VARCHAR(255) NOT NULL,
    room_id VARCHAR(255) NOT NULL,
    check_in TIMESTAMP WITH TIME ZONE NOT NULL,
    check_out TIMESTAMP WITH TIME ZONE NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    total_amount_cents BIGINT NOT NULL,
    total_amount_currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    cancellation_reason TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT valid_reservation_status CHECK (status IN ('pending', 'confirmed', 'active', 'completed', 'cancelled')),
    CONSTRAINT valid_date_range CHECK (check_out > check_in)
);

-- ======================================
-- Reservation Guests Table
-- ======================================
-- Stores GuestInfo entities within the Reservation aggregate.
-- One reservation can have multiple guests.
CREATE TABLE IF NOT EXISTS reservation_guests (
    id SERIAL PRIMARY KEY,
    reservation_id VARCHAR(255) NOT NULL REFERENCES reservations(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    phone_number VARCHAR(50)
);

-- ======================================
-- Indexes for Reservation Queries
-- ======================================

-- Reservation lookups by guest (for listing user's reservations)
CREATE INDEX IF NOT EXISTS idx_reservations_guest_id ON reservations(guest_id);

-- Reservation lookups by room (for availability checking)
CREATE INDEX IF NOT EXISTS idx_reservations_room_id ON reservations(room_id);

-- Filter reservations by status
CREATE INDEX IF NOT EXISTS idx_reservations_status ON reservations(status);

-- Date range queries for availability checking
CREATE INDEX IF NOT EXISTS idx_reservations_check_in ON reservations(check_in);
CREATE INDEX IF NOT EXISTS idx_reservations_check_out ON reservations(check_out);

-- Composite index for availability queries (room + date range)
CREATE INDEX IF NOT EXISTS idx_reservations_room_dates ON reservations(room_id, check_in, check_out);

-- Guest lookups by reservation
CREATE INDEX IF NOT EXISTS idx_reservation_guests_reservation_id ON reservation_guests(reservation_id);
