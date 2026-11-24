CREATE TABLE public.homeworks (
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    student_id  BIGINT NOT NULL       REFERENCES public.students(id) ON DELETE RESTRICT,
    version     BIGINT NOT NULL  DEFAULT 1,
    title       TEXT NOT NULL    CHECK(char_length(title) BETWEEN 3 AND 30),
    description TEXT NOT NULL    CHECK(char_length(description) BETWEEN 3 AND 3500),
    due_at      TIMESTAMPTZ,  
    created_at  TIMESTAMPTZ NOT NULL   DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX homeworks_student_id_idx 
ON public.homeworks(
    student_id
);