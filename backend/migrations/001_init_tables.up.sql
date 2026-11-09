CREATE TABLE IF NOT EXISTS notification_preferences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tg_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    email_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    tg_chat_id BIGINT
);

CREATE TABLE IF NOT EXISTS clients (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    id_notification_preferences UUID NOT NULL,
    email VARCHAR NOT NULL UNIQUE,
    firstname VARCHAR NOT NULL,
    lastname VARCHAR NOT NULL,
    surname VARCHAR NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_visit_at TIMESTAMPTZ,

    FOREIGN KEY (id_notification_preferences) REFERENCES notification_preferences(id)
);

CREATE TABLE IF NOT EXISTS masters (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    id_notification_preferences UUID NOT NULL,
    email VARCHAR NOT NULL UNIQUE,
    firstname VARCHAR NOT NULL,
    lastname VARCHAR NOT NULL,
    surname VARCHAR NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    FOREIGN KEY (id_notification_preferences) REFERENCES notification_preferences(id)
);

CREATE TABLE IF NOT EXISTS appointments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    id_user UUID NOT NULL,
    title VARCHAR NOT NULL,
    description VARCHAR NOT NULL,
    price_penny BIGINT NOT NULL,
    duration_minutes INTEGER NOT NULL DEFAULT 0,
    interval_minutes INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    archived_at TIMESTAMPTZ,

    FOREIGN KEY (id_user) REFERENCES masters(id)
);

CREATE TABLE IF NOT EXISTS appointment_booking_statuses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    status VARCHAR NOT NULL
);

CREATE TABLE IF NOT EXISTS appointment_bookings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    id_client UUID NOT NULL,
    id_appointment UUID NOT NULL,
    id_master UUID NOT NULL,
    id_status UUID NOT NULL,
    comment VARCHAR(512),
    fixed_price_penny BIGINT NOT NULL DEFAULT 0,
    start_at TIMESTAMPTZ NOT NULL,
    end_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    FOREIGN KEY (id_client) REFERENCES clients(id),
    FOREIGN KEY (id_appointment) REFERENCES appointments(id),
    FOREIGN KEY (id_master) REFERENCES masters(id),
    FOREIGN KEY (id_status) REFERENCES appointment_booking_statuses(id)
);

CREATE TABLE IF NOT EXISTS master_schedules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    id_user UUID NOT NULL,
    weekday INTEGER NOT NULL DEFAULT 1,
    breaks JSONB NOT NULL DEFAULT '[]'::JSONB,
    start_at TIME NOT NULL,
    end_at TIME NOT NULL,

    FOREIGN KEY (id_user) REFERENCES masters(id)
);

CREATE TABLE IF NOT EXISTS tariffs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR NOT NULL,
    price_penny BIGINT NOT NULL DEFAULT 0,
    days INTEGER NOT NULL DEFAULT 0,
    benefits VARCHAR NOT NULL
);

CREATE TABLE IF NOT EXISTS master_tariff (
    id_user UUID PRIMARY KEY,
    id_current_tariff UUID,
    paid_until TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_paid TIMESTAMPTZ,

    FOREIGN KEY (id_user) REFERENCES masters(id),
    FOREIGN KEY (id_current_tariff) REFERENCES tariffs(id)
);

CREATE TABLE IF NOT EXISTS master_payment_statuses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    status VARCHAR NOT NULL
);

CREATE TABLE IF NOT EXISTS master_payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    id_user UUID NOT NULL,
    id_status UUID NOT NULL,
    id_yookassa UUID NOT NULL,
    price_penny BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    confirmed_at TIMESTAMPTZ,

    FOREIGN KEY (id_user) REFERENCES masters(id),
    FOREIGN KEY (id_status) REFERENCES master_payment_statuses(id)
);
