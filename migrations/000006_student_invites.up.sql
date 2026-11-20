CREATE TABLE public.student_invites (
    id                    BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    token_hash            TEXT NOT NULL UNIQUE,
    student_name          TEXT NOT NULL,
    requester_telegram_id BIGINT,
    status                TEXT NOT NULL CHECK(status IN ('issued', 'pending',
    'approved', 'rejected')) DEFAULT 'issued',
    expires_at            TIMESTAMPTZ NOT NULL,
    requested_at          TIMESTAMPTZ,
    decided_at            TIMESTAMPTZ,
    created_at            TIMESTAMPTZ NOT NULL    DEFAULT CURRENT_TIMESTAMP
);