-- ======================================
-- Payment Domain Schema
-- ======================================
-- Schema for the Payment bounded context.
-- This script runs automatically on first PostgreSQL startup.

-- ======================================
-- Payments Table
-- ======================================
-- Stores the Payment aggregate root.
-- Status transitions: pending -> authorized -> captured (or failed/refunded)
-- Note: reservation_id is NOT a foreign key since reservations are in a separate database.
-- Cross-database referential integrity is maintained via domain events.
CREATE TABLE IF NOT EXISTS payments (
    id VARCHAR(255) PRIMARY KEY,
    reservation_id VARCHAR(255) NOT NULL,
    amount_cents BIGINT NOT NULL,
    amount_currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    payment_method VARCHAR(100),
    transaction_id VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT valid_payment_status CHECK (status IN ('pending', 'authorized', 'captured', 'failed', 'refunded'))
);

-- ======================================
-- Payment Attempts Table
-- ======================================
-- Stores PaymentAttempt entities within the Payment aggregate.
-- Tracks history of payment processing attempts.
CREATE TABLE IF NOT EXISTS payment_attempts (
    id SERIAL PRIMARY KEY,
    payment_id VARCHAR(255) NOT NULL REFERENCES payments(id) ON DELETE CASCADE,
    attempted_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    status VARCHAR(50) NOT NULL,
    error_code VARCHAR(100),
    error_msg TEXT
);

-- ======================================
-- Indexes for Payment Queries
-- ======================================

-- Payment lookups by reservation
CREATE INDEX IF NOT EXISTS idx_payments_reservation_id ON payments(reservation_id);

-- Filter payments by status
CREATE INDEX IF NOT EXISTS idx_payments_status ON payments(status);

-- Payment attempts by payment
CREATE INDEX IF NOT EXISTS idx_payment_attempts_payment_id ON payment_attempts(payment_id);
