ALTER TABLE public.lessons
ADD version BIGINT NOT NULL DEFAULT 1
CHECK (version >= 1);