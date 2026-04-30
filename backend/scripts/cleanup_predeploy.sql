-- Pre-deploy cleanup script for production cutover.
-- Purpose:
-- 1) Remove test/demo transactional data.
-- 2) Remove teacher/student accounts.
-- 3) Keep admin accounts.
--
-- Run from backend folder:
--   psql "$DATABASE_URL" -f scripts/cleanup_predeploy.sql
--
-- IMPORTANT:
-- - Backup database first.
-- - This is destructive.

BEGIN;

-- Auth/log artifacts
DELETE FROM password_reset_tokens;
DELETE FROM login_logs;
DELETE FROM audit_logs;

-- Student runtime artifacts
DELETE FROM student_exam_dismissals;
DELETE FROM student_attendance;
DELETE FROM attendance_sessions;

-- Exam runtime artifacts
DELETE FROM exam_events;
DELETE FROM exam_attempts;
DELETE FROM exam_session_questions;
DELETE FROM exam_sessions;

-- LTI runtime/user-linked artifacts
DELETE FROM lti_ags_launches;
DELETE FROM lti_sessions;
DELETE FROM lti_users;
DELETE FROM lti_nonces;

-- Exam definitions
DELETE FROM exam_tokens;
DELETE FROM exam_targets;
DELETE FROM exam_question_sets;
DELETE FROM exams;

-- Question bank data
DELETE FROM question_true_false_statements;
DELETE FROM question_true_false;
DELETE FROM question_essays;
DELETE FROM question_short_answers;
DELETE FROM question_matching_pairs;
DELETE FROM question_options;
DELETE FROM questions;
DELETE FROM question_sets;

-- Content data
DELETE FROM announcements;
DELETE FROM registration_requests;

-- Teacher/student assignments
DELETE FROM teacher_subjects;
DELETE FROM teacher_groups;
DELETE FROM teacher_levels;

-- Master accounts (keep admin users)
DELETE FROM teachers;
DELETE FROM students;
DELETE FROM users WHERE role IN ('teacher', 'student');

COMMIT;

-- Post-check summary
SELECT
  (SELECT COUNT(*) FROM users WHERE role = 'admin') AS admin_users,
  (SELECT COUNT(*) FROM users WHERE role = 'teacher') AS teacher_users,
  (SELECT COUNT(*) FROM users WHERE role = 'student') AS student_users,
  (SELECT COUNT(*) FROM exams) AS exams,
  (SELECT COUNT(*) FROM question_sets) AS question_sets,
  (SELECT COUNT(*) FROM exam_sessions) AS exam_sessions,
  (SELECT COUNT(*) FROM exam_attempts) AS exam_attempts;

