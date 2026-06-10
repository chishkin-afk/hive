-- USER'S TABLE
create table
    if not exists users (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
        firstname VARCHAR(128) NOT NULL,
        lastname VARCHAR(128),
        confirmed_email VARCHAR(256) NOT NULL,
        password_hash VARCHAR(255) NOT NULL,
        paid_until TIMESTAMPTZ NOT NULL DEFAULT NOW (),
        on_vacation BOOLEAN DEFAULT FALSE NOT NULL,
        tg_chat_id BIGINT,
        notification JSONB NOT NULL,
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW (),
        updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW ()
    );

create index if not exists idx_user_email on users (confirmed_email)
where
    confirmed_email is not null;

create index if not exists idx_users_tg_chat_id on users (tg_chat_id)
where
    tg_chat_id is not null;

-- OFFER'S TABLE
create table
    if not exists offers (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
        author_id UUID NOT NULL,
        title VARCHAR(128) NOT NULL,
        description VARCHAR(512),
        price BIGINT NOT NULL DEFAULT 0,
        slot_duration_minutes BIGINT NOT NULL DEFAULT 0,
        slot_interval_minutes BIGINT NOT NULL DEFAULT 15,
        is_active BOOLEAN DEFAULT TRUE,
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW (),
        updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW (),
        deleted_at TIMESTAMPTZ DEFAULT,
        FOREIGN KEY (author_id) REFERENCES users (id)
    );

-- SCHEDULE'S TABLE
create table
    if not exists schedules (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
        user_id UUID NOT NULL,
        day_of_week BIGINT NOT NULL DEFAULT 1,
        is_working_day BOOLEAN DEFAULT TRUE,
        working_time JSONB NOT NULL,
        breaks JSONB,
        FOREIGN KEY (user_id) REFERENCES users (id),
        CONSTRAINT unique_user_day UNIQUE (user_id, day_of_week)
    );

-- BOOKING'S TABLE
create table
    if not exists bookings (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
        offer_id UUID NOT NULL,
        user_id UUID NOT NULL,
        comment VARCHAR(128),
        start_time TIMESTAMPTZ NOT NULL,
        is_cancelled BOOLEAN NOT NULL DEFAULT FALSE,
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW (),
        updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW (),
        FOREIGN KEY (offer_id) REFERENCES offers (id),
        FOREIGN KEY (user_id) REFERENCES users (id)
    );

CREATE UNIQUE INDEX if not exists  idx_unique_offer_slot ON bookings (offer_id, start_time) WHERE is_cancelled = FALSE;

-- CLIENT'S TABLE
create table
    if not exists clients (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
        user_id UUID NOT NULL,
        offer_id UUID NOT NULL,
        visits BIGINT NOT NULL DEFAULT 0,
        last_visit TIMESTAMPTZ NOT NULL,
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW (),
        FOREIGN KEY (user_id) REFERENCES users (id),
        FOREIGN KEY (offer_id) REFERENCES offers (id),
        CONSTRAINT unique_client UNIQUE (user_id, offer_id)
    );

create or replace function check_offer_active()
return trigger as $$
begin
    if NEW.offer_id is not null then
        if not exists(
            select 1 from offers where id = NEW.offer_id and deleted_at is null
        ) then
            raise EXCEPTION 'Referential integrity violation: offer (id: %)', NEW.offer_id
            using ERRCODE = 'foreign_key_violation';
        end if;
    end if;
    
    return NEW;
end;
$$ language plpgsql;

create trigger trg_ensure_offer_active
before insert or update of offer_id on clients
for EACH row
execute function check_offer_active();

-- PAYMENT'S HISTORY
create table
    if not exists payments_history (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
        payment_id UUID NOT NULL,
        user_id UUID NOT NULL,
        amount BIGINT NOT NULL,
        is_confirmed BOOLEAN DEFAULT FALSE,
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW (),
        confirmed_at TIMESTAMPTZ,
        FOREIGN KEY (user_id) REFERENCES users (id)
    );