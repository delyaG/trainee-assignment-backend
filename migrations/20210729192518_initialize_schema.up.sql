-- Build-in PostgreSQL extension to update timestamp fields with moddatetime procedure;
-- In our case is used to set actual 'updated' before UPDATE.
CREATE EXTENSION IF NOT EXISTS moddatetime;

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Status meaning:
-- 00011111
-- │││││││└ user has started a registration process
-- ││││││└─ user has confirmed a phone
-- │││││└── user has finished a registration process
-- ││││└─── user has saved an email
-- │││└──── user has confirmed an email
-- ││└───── reserved for future use
-- │└────── reserved for future use
-- └─────── reserved for future use
CREATE TABLE IF NOT EXISTS users
(
    id          INTEGER GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
    status      BIT VARYING NOT NULL DEFAULT B'00000000',
    phone       TEXT        NOT NULL UNIQUE,
    first_name  TEXT,
    middle_name TEXT,
    last_name   TEXT,
    birthday    DATE,
    city        TEXT,
    email       TEXT UNIQUE,
    created_at  TIMESTAMP            DEFAULT now() NOT NULL,
    updated_at  TIMESTAMP
);

CREATE TRIGGER update_updated_at
    BEFORE UPDATE
    ON users
    FOR EACH ROW
EXECUTE PROCEDURE moddatetime(updated_at);

CREATE TABLE IF NOT EXISTS refresh_sessions
(
    id            INTEGER GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
    user_id       INTEGER REFERENCES users (id) ON UPDATE CASCADE ON DELETE CASCADE,
    refresh_token UUID               DEFAULT uuid_generate_v4() NOT NULL,
    fingerprint   TEXT      NOT NULL,
    user_agent    TEXT      NOT NULL,
    ip            TEXT      NOT NULL,
    expires_at    TIMESTAMP NOT NULL,
    created_at    TIMESTAMP NOT NULL DEFAULT now()
);