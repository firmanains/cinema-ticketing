-- ============================================================
-- Sistem Pemesanan Tiket Bioskop
-- Database: PostgreSQL
-- ============================================================

-- Extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================================
-- GROUP 1: Cinema & Scheduling
-- ============================================================

CREATE TABLE cinemas (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name       VARCHAR(100) NOT NULL,
    city       VARCHAR(50)  NOT NULL,
    address    TEXT         NOT NULL,
    created_at TIMESTAMP    NOT NULL DEFAULT NOW()
);

CREATE TABLE studios (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    cinema_id   UUID         NOT NULL REFERENCES cinemas(id) ON DELETE CASCADE,
    name        VARCHAR(50)  NOT NULL,
    total_seats INT          NOT NULL CHECK (total_seats > 0),
    created_at  TIMESTAMP    NOT NULL DEFAULT NOW()
);

CREATE TABLE seats (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    studio_id  UUID        NOT NULL REFERENCES studios(id) ON DELETE CASCADE,
    row        CHAR(1)     NOT NULL,
    number     INT         NOT NULL CHECK (number > 0),
    type       VARCHAR(20) NOT NULL DEFAULT 'regular'
                           CHECK (type IN ('regular', 'vip', 'couple')),
    created_at TIMESTAMP   NOT NULL DEFAULT NOW(),
    UNIQUE (studio_id, row, number)
);

CREATE TABLE movies (
    id               UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title            VARCHAR(150) NOT NULL,
    duration_minutes INT          NOT NULL CHECK (duration_minutes > 0),
    genre            VARCHAR(50),
    rating           VARCHAR(10),
    created_at       TIMESTAMP    NOT NULL DEFAULT NOW()
);

CREATE TABLE showtimes (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    studio_id  UUID           NOT NULL REFERENCES studios(id) ON DELETE RESTRICT,
    movie_id   UUID           NOT NULL REFERENCES movies(id)  ON DELETE RESTRICT,
    start_time TIMESTAMP      NOT NULL,
    price      NUMERIC(12, 2) NOT NULL CHECK (price >= 0),
    status     VARCHAR(20)    NOT NULL DEFAULT 'scheduled'
                              CHECK (status IN ('scheduled', 'ongoing', 'done', 'cancelled')),
    created_at TIMESTAMP      NOT NULL DEFAULT NOW()
);

-- ============================================================
-- GROUP 2: Booking & Payment
-- ============================================================

CREATE TABLE users (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name       VARCHAR(100) NOT NULL,
    email      VARCHAR(150) NOT NULL UNIQUE,
    phone      VARCHAR(20),
    created_at TIMESTAMP    NOT NULL DEFAULT NOW()
);

CREATE TABLE bookings (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id     UUID        NOT NULL REFERENCES users(id)     ON DELETE RESTRICT,
    showtime_id UUID        NOT NULL REFERENCES showtimes(id) ON DELETE RESTRICT,
    status      VARCHAR(20) NOT NULL DEFAULT 'pending'
                            CHECK (status IN ('pending', 'confirmed', 'cancelled')),
    expires_at  TIMESTAMP   NOT NULL DEFAULT NOW() + INTERVAL '5 minutes',
    created_at  TIMESTAMP   NOT NULL DEFAULT NOW()
);

CREATE TABLE booking_seats (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    booking_id UUID      NOT NULL REFERENCES bookings(id) ON DELETE CASCADE,
    seat_id    UUID      NOT NULL REFERENCES seats(id)    ON DELETE RESTRICT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE (booking_id, seat_id)
);

CREATE TABLE payments (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    booking_id  UUID           NOT NULL UNIQUE REFERENCES bookings(id) ON DELETE RESTRICT,
    method      VARCHAR(30)    NOT NULL CHECK (method IN ('credit_card', 'debit_card', 'e_wallet', 'bank_transfer', 'qris')),
    status      VARCHAR(20)    NOT NULL DEFAULT 'pending'
                               CHECK (status IN ('pending', 'paid', 'failed', 'refunded')),
    gateway_ref VARCHAR(100),
    amount      NUMERIC(12, 2) NOT NULL CHECK (amount > 0),
    paid_at     TIMESTAMP,
    created_at  TIMESTAMP      NOT NULL DEFAULT NOW()
);

CREATE TABLE refunds (
    id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    booking_id   UUID           NOT NULL REFERENCES bookings(id) ON DELETE RESTRICT,
    initiated_by VARCHAR(10)    NOT NULL CHECK (initiated_by IN ('user', 'cinema')),
    amount       NUMERIC(12, 2) NOT NULL CHECK (amount > 0),
    status       VARCHAR(20)    NOT NULL DEFAULT 'pending'
                                CHECK (status IN ('pending', 'processed', 'failed')),
    created_at   TIMESTAMP      NOT NULL DEFAULT NOW()
);

-- ============================================================
-- INDEXES
-- ============================================================

-- Showtimes
CREATE INDEX idx_showtimes_studio_id  ON showtimes (studio_id);
CREATE INDEX idx_showtimes_movie_id   ON showtimes (movie_id);
CREATE INDEX idx_showtimes_start_time ON showtimes (start_time);
CREATE INDEX idx_showtimes_status     ON showtimes (status);

-- Bookings
CREATE INDEX idx_bookings_user_id     ON bookings (user_id);
CREATE INDEX idx_bookings_showtime_id ON bookings (showtime_id);
CREATE INDEX idx_bookings_status      ON bookings (status);
CREATE INDEX idx_bookings_expires_at  ON bookings (expires_at) WHERE status = 'pending';

-- Booking seats
CREATE INDEX idx_booking_seats_booking_id ON booking_seats (booking_id);
CREATE INDEX idx_booking_seats_seat_id    ON booking_seats (seat_id);

-- Payments
CREATE INDEX idx_payments_booking_id  ON payments (booking_id);
CREATE INDEX idx_payments_status      ON payments (status);

-- Refunds
CREATE INDEX idx_refunds_booking_id ON refunds (booking_id);
CREATE INDEX idx_refunds_status     ON refunds (status);

-- ============================================================
-- COMMENTS
-- ============================================================

COMMENT ON COLUMN bookings.expires_at  IS 'Sinkron dengan Redis TTL (5 menit). Digunakan untuk cleanup booking pending yang kedaluwarsa.';
COMMENT ON COLUMN bookings.status      IS 'Alur: pending -> confirmed -> cancelled';
COMMENT ON COLUMN payments.gateway_ref IS 'Reference ID dari payment gateway (Midtrans/Xendit) untuk rekonsiliasi.';
COMMENT ON COLUMN refunds.initiated_by IS 'Membedakan sumber pembatalan: user (mandiri) atau cinema (showtime cancelled).';
