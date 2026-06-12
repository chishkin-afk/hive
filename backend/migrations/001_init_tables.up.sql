-- USERS TABLE
CREATE TABLE IF NOT EXISTS public.users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    firstname VARCHAR(128) NOT NULL,
    lastname VARCHAR(128),
    confirmed_email VARCHAR(256) NOT NULL,
    phone_number VARCHAR(128),
    password_hash VARCHAR(255) NOT NULL,
    paid_until TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    on_vacation BOOLEAN NOT NULL DEFAULT FALSE,
    tg_chat_id BIGINT,
    notification JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_user_email ON public.users (confirmed_email) WHERE confirmed_email IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_users_tg_chat_id ON public.users (tg_chat_id) WHERE tg_chat_id IS NOT NULL;

-- OFFERS TABLE
CREATE TABLE IF NOT EXISTS public.offers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    author_id UUID NOT NULL,
    title VARCHAR(128) NOT NULL,
    description VARCHAR(512),
    price BIGINT NOT NULL DEFAULT 0,
    slot_duration_minutes INTEGER NOT NULL DEFAULT 0, 
    slot_interval_minutes INTEGER NOT NULL DEFAULT 15,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT fk_offers_author FOREIGN KEY (author_id) REFERENCES public.users (id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_offers_author_id ON public.offers (author_id);

-- SCHEDULES TABLE
CREATE TABLE IF NOT EXISTS public.schedules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    day_of_week SMALLINT NOT NULL DEFAULT 1,
    is_working_day BOOLEAN DEFAULT TRUE,
    working_time JSONB NOT NULL DEFAULT '{}'::jsonb,
    breaks JSONB,
    CONSTRAINT fk_schedules_user FOREIGN KEY (user_id) REFERENCES public.users (id) ON DELETE CASCADE,
    CONSTRAINT unique_user_day UNIQUE (user_id, day_of_week)
);

-- BOOKINGS TABLE
CREATE TABLE IF NOT EXISTS public.bookings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    offer_id UUID NOT NULL,
    user_id UUID NOT NULL,
    comment VARCHAR(128),
    start_time TIMESTAMPTZ NOT NULL,
    is_cancelled BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_bookings_offer FOREIGN KEY (offer_id) REFERENCES public.offers (id),
    CONSTRAINT fk_bookings_user FOREIGN KEY (user_id) REFERENCES public.users (id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_unique_offer_slot ON public.bookings (offer_id, start_time) WHERE is_cancelled = FALSE;

-- CLIENTS TABLE
CREATE TABLE IF NOT EXISTS public.clients (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    offer_id UUID NOT NULL,
    visits INTEGER NOT NULL DEFAULT 0,
    last_visit TIMESTAMPTZ NOT NULL DEFAULT NOW(), 
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_clients_user FOREIGN KEY (user_id) REFERENCES public.users (id) ON DELETE CASCADE,
    CONSTRAINT fk_clients_offer FOREIGN KEY (offer_id) REFERENCES public.offers (id),
    CONSTRAINT unique_client UNIQUE (user_id, offer_id)
);

-- PAYMENTS HISTORY TABLE
CREATE TABLE IF NOT EXISTS public.payments_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    payment_id UUID NOT NULL, 
    user_id UUID NOT NULL,
    amount BIGINT NOT NULL,
    is_confirmed BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    confirmed_at TIMESTAMPTZ,
    CONSTRAINT fk_payments_user FOREIGN KEY (user_id) REFERENCES public.users (id) ON DELETE CASCADE
);

-- TRIGGER: ENSURE OFFER ACTIVE (FIXED & IDEMPOTENT)
DROP TRIGGER IF EXISTS trg_ensure_offer_active_clients ON public.clients;
DROP TRIGGER IF EXISTS trg_ensure_offer_active_bookings ON public.bookings;

DROP FUNCTION IF EXISTS public.check_offer_active();

CREATE OR REPLACE FUNCTION public.check_offer_active()
RETURNS trigger AS $$
BEGIN
    IF NEW.offer_id IS NOT NULL THEN
        IF NOT EXISTS (
            SELECT 1 
            FROM public.offers 
            WHERE id = NEW.offer_id 
              AND deleted_at IS NULL
            FOR KEY SHARE
        ) THEN
            RAISE EXCEPTION 'Referential integrity violation: offer (id: %) is missing or soft-deleted', NEW.offer_id
            USING ERRCODE = 'foreign_key_violation';
        END IF;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

CREATE TRIGGER trg_ensure_offer_active_clients
BEFORE INSERT OR UPDATE OF offer_id ON public.clients
FOR EACH ROW
EXECUTE FUNCTION public.check_offer_active();

CREATE TRIGGER trg_ensure_offer_active_bookings
BEFORE INSERT OR UPDATE OF offer_id ON public.bookings
FOR EACH ROW
EXECUTE FUNCTION public.check_offer_active();