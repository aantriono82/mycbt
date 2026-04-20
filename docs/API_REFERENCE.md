# AtigaCBT API Reference

This document provides a comprehensive reference for the AtigaCBT API.

## Base URL
`https://your-domain.com/api/v1`

---

## 1. Authentication & Profile
- `POST /auth/login`
  - Payload: `{ "username": "...", "password": "..." }`
  - Response: `{ "data": { "token": "...", "user": { ... } } }`
- `GET /me`
  - Auth required. Returns current user profile and role.
- `POST /registrations` (Public)
  - Student self-registration.
  - Payload: `{ "username": "...", "name": "...", "nis": "...", "password": "...", "jenjang": "...", "phone": "..." }`

---

## 2. Master Data (Admin Only)
- `GET /admin/teachers` - List teachers.
- `POST /admin/teachers/import` - Bulk import via Excel.
- `GET /admin/students` - List students.
- `GET /lookups/levels` - List educational levels.
- `GET /lookups/groups` - List classes/groups.

---

## 3. Question Bank (Admin/Teacher)
- `GET /question-sets` - List question sets.
- `POST /question-sets` - Create set.
- `POST /question-sets/:id/questions` - Create question.
  - Supports: `mc_single`, `mc_multiple`, `matching`, `short_answer`, `essay`, `true_false`.
- `POST /question-sets/:id/import-docx` - Import from Word document.

---

## 4. Exams & Scheduling
- `POST /exams` - Create exam schedule.
- `PUT /exams/:id/targets` - Set target participants (Levels, Groups, or specific Students).
  - Note: Teachers are restricted to their assigned groups/levels.
- `POST /exams/:id/tokens` - Generate exam token.
- `GET /exams/:id/monitor/sessions` - Real-time monitoring of student sessions.

---

## 5. Student Exam Room
- `GET /student/exams` - List available exams for the student.
- `POST /student/exams/:id/join`
  - Payload: `{ "token": "..." }`
  - Response: `{ "data": { "session_id": "..." } }`
- `GET /student/sessions/:id/questions` - Fetch questions for an active session.
- `POST /student/sessions/:id/answers` - Save/Update answer (autosave).
- `POST /student/sessions/:id/submit` - Final submission.
- `POST /student/sessions/:id/heartbeat` - Maintain session & detect focus loss.

---

## 6. Analytics & Results
- `GET /exams/:id/results` - List exam results.
- `GET /exams/:id/item-analysis` - Statistical data (P-Value, D-Index).
- `GET /exams/:id/item-analysis/suggestions` - AI-powered item improvement tips.
- `GET /exams/:id/export` - Comprehensive Excel report export.
- `POST /exams/:id/results/blast` - Send results to students via Email/WhatsApp.

---

## 7. LTI 1.3 Interoperability
- `GET /lti/login` - OIDC initiation.
- `POST /lti/launch` - LTI resource link launch.
- `POST /lti/deep-link` - Resource selection for LMS.

---

## 8. Maintenance
- `GET /maintenance/backup` - Download database SQL dump.
- `POST /maintenance/restore` - Restore from SQL dump.
