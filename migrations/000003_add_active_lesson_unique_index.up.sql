CREATE UNIQUE INDEX lessons_active_student_datetime_uidx 
ON public.lessons(
    student_id, date_time
)
WHERE status IN ('scheduled', 'in_progress');