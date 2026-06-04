-- Migration 009: Sync grades from student_submissions to assistant_reports
-- This migration creates a trigger to sync grades when student_submissions are updated

-- Create trigger function to sync grades
CREATE OR REPLACE FUNCTION sync_submission_grade_to_report()
RETURNS TRIGGER AS $$
BEGIN
  -- Update assistant_reports with the score from student_submissions
  UPDATE assistant_reports
  SET score = NEW.score, feedback = NEW.feedback
  WHERE nim = (SELECT student_id FROM students WHERE id = NEW.student_id LIMIT 1)
    AND week = (SELECT session_number FROM course_sessions WHERE id = (SELECT session_id FROM session_assignments WHERE id = NEW.assignment_id LIMIT 1) LIMIT 1)
    AND course_code = (SELECT class_code FROM courses WHERE id = (SELECT course_id FROM session_assignments WHERE id = NEW.assignment_id LIMIT 1) LIMIT 1);
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Drop existing trigger if it exists
DROP TRIGGER IF EXISTS sync_submission_grade_trigger ON student_submissions;

-- Create trigger
CREATE TRIGGER sync_submission_grade_trigger
AFTER UPDATE ON student_submissions
FOR EACH ROW
WHEN (NEW.score IS DISTINCT FROM OLD.score OR NEW.feedback IS DISTINCT FROM OLD.feedback)
EXECUTE FUNCTION sync_submission_grade_to_report();
