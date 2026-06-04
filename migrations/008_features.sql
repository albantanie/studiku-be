-- Migration 008: sort_order for course_sessions, rejection_note for assistant_reports, file upload support
-- Safe to run multiple times (IF NOT EXISTS / ADD COLUMN IF NOT EXISTS)

-- 1. Add sort_order to course_sessions
ALTER TABLE course_sessions ADD COLUMN IF NOT EXISTS sort_order INT NOT NULL DEFAULT 0;
-- Backfill sort_order from session_number for existing rows
UPDATE course_sessions SET sort_order = session_number WHERE sort_order = 0;

-- 2. Add rejection_note to assistant_reports
ALTER TABLE assistant_reports ADD COLUMN IF NOT EXISTS rejection_note TEXT NOT NULL DEFAULT '';

-- 3. Add file_url to course_materials (for preview)
ALTER TABLE course_materials ADD COLUMN IF NOT EXISTS file_url TEXT NOT NULL DEFAULT '';

-- 4. Add file_url to session_assignments (for assignment file attachment)
ALTER TABLE session_assignments ADD COLUMN IF NOT EXISTS file_url TEXT NOT NULL DEFAULT '';

-- 5. Add assistant_attendance table for per-session aslab attendance
CREATE TABLE IF NOT EXISTS assistant_session_attendance (
  id SERIAL PRIMARY KEY,
  session_id INT NOT NULL REFERENCES course_sessions(id) ON DELETE CASCADE,
  status VARCHAR(20) NOT NULL DEFAULT 'Hadir',
  check_in_time VARCHAR(20) NOT NULL DEFAULT '',
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(session_id)
);
