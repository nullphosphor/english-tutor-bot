CREATE TABLE public.students (
    id               BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    telegram_id      BIGINT UNIQUE NOT NULL,
    is_active        BOOLEAN NOT NULL                    DEFAULT TRUE
);