CREATE TABLE public.lessons (
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    student_id  BIGINT NOT NULL       REFERENCES public.students(id) ON DELETE RESTRICT,
    date_time   TIMESTAMPTZ NOT NULL,
    started_at  TIMESTAMPTZ,       
    finished_at TIMESTAMPTZ,     
    meeting_app TEXT NOT NULL,
    meeting_url TEXT NOT NULL,
    status      TEXT NOT NULL          DEFAULT 'scheduled',
    created_at  TIMESTAMPTZ NOT NULL   DEFAULT CURRENT_TIMESTAMP
);