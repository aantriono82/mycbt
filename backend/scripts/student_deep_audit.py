#!/usr/bin/env python3
import datetime
import json
import sys
import time
import urllib.error
import urllib.request

BASE = "http://127.0.0.1:8080"


def req(method, path, token=None, body=None):
    data = None
    headers = {}
    if token:
        headers["Authorization"] = f"Bearer {token}"
    if body is not None:
        headers["Content-Type"] = "application/json"
        data = json.dumps(body).encode()
    request = urllib.request.Request(
        BASE + path, data=data, method=method, headers=headers
    )
    try:
        with urllib.request.urlopen(request, timeout=30) as response:
            raw = response.read().decode()
            payload = json.loads(raw) if raw else {}
            return response.status, payload
    except urllib.error.HTTPError as err:
        raw = err.read().decode()
        try:
            payload = json.loads(raw) if raw else {}
        except json.JSONDecodeError:
            payload = {"_raw": raw}
        return err.code, payload


def expect(checks, name, status, allowed, body=None, note=""):
    ok = status in allowed
    err = body.get("error") if isinstance(body, dict) else None
    checks.append((name, status, sorted(allowed), ok, err, note))


def fatal(message, status, body):
    print(f"FATAL {message}: status={status} body={body}")
    sys.exit(1)


def require_success(message, status, body):
    if status not in (200, 201):
        fatal(message, status, body)


def main():
    checks = []
    run_id = str(int(time.time()))

    status, login = req(
        "POST", "/api/v1/auth/login", body={"username": "admin", "password": "admin12345"}
    )
    require_success("admin login", status, login)
    admin = login["data"]["access_token"]

    # Stabilize system config to deterministic student join behavior.
    status, system = req("GET", "/api/v1/settings/system", admin)
    require_success("get system settings", status, system)
    sys_data = system.get("data") or {}
    put_system = {
        "timezone": sys_data.get("timezone") or "Asia/Jakarta",
        "token_required": True,
        "allow_reset_login": bool(sys_data.get("allow_reset_login")),
        "max_active_sessions": int(sys_data.get("max_active_sessions") or 1),
        "attendance_require_ip": bool(sys_data.get("attendance_require_ip")),
    }
    status, body = req("PUT", "/api/v1/settings/system", admin, put_system)
    require_success("put system settings", status, body)

    # Fixture master data
    _, program = req(
        "POST", "/api/v1/admin/programs", admin, {"code": f"PRG-S-{run_id}", "name": f"Program S {run_id}"}
    )
    program_id = (program.get("data") or {}).get("id")

    _, level = req(
        "POST", "/api/v1/admin/levels", admin, {"name": f"LevelS {run_id}", "kelas": 9}
    )
    level_id = (level.get("data") or {}).get("id")

    _, group = req(
        "POST", "/api/v1/admin/groups", admin, {"name": f"GroupS {run_id}"}
    )
    group_id = (group.get("data") or {}).get("id")

    subject_code = f"SBS{run_id}"
    _, subject = req(
        "POST", "/api/v1/admin/subjects", admin, {"code": subject_code, "name": f"Subject S {run_id}"}
    )
    subject_id = (subject.get("data") or {}).get("id")

    _, session = req(
        "POST", "/api/v1/admin/sessions", admin, {"name": f"Sesi S {run_id}", "start_time": "00:00", "end_time": "23:59"}
    )
    session_id = (session.get("data") or {}).get("id")

    teacher_username = f"guruS_{run_id}"
    teacher_password = "Guru12345!"
    _, teacher = req(
        "POST",
        "/api/v1/admin/teachers",
        admin,
        {
            "username": teacher_username,
            "password": teacher_password,
            "name": f"Guru S {run_id}",
            "email": f"{teacher_username}@x.test",
            "nip": f"{run_id}9",
            "jenjang": "SMP",
            "mapel_codes": subject_code,
            "group_names": f"GroupS {run_id}",
            "level_names": f"LevelS {run_id}",
        },
    )
    teacher_id = (teacher.get("data") or {}).get("id")

    student1_username = f"siswaS1_{run_id}"
    student2_username = f"siswaS2_{run_id}"
    student_password = "Siswa12345!"

    _, s1 = req(
        "POST",
        "/api/v1/admin/students",
        admin,
        {
            "username": student1_username,
            "password": student_password,
            "name": f"Siswa S1 {run_id}",
            "email": f"{student1_username}@x.test",
            "nis": f"{run_id}1",
            "jenjang": "SMP",
            "program_id": program_id,
            "level_id": level_id,
            "group_id": group_id,
        },
    )
    student1_id = (s1.get("data") or {}).get("id")

    _, s2 = req(
        "POST",
        "/api/v1/admin/students",
        admin,
        {
            "username": student2_username,
            "password": student_password,
            "name": f"Siswa S2 {run_id}",
            "email": f"{student2_username}@x.test",
            "nis": f"{run_id}2",
            "jenjang": "SMP",
            "program_id": program_id,
            "level_id": level_id,
            "group_id": group_id,
        },
    )
    student2_id = (s2.get("data") or {}).get("id")

    if not all([program_id, level_id, group_id, subject_id, session_id, teacher_id, student1_id, student2_id]):
        fatal("fixture create ids missing", 500, {"program_id": program_id, "level_id": level_id, "group_id": group_id, "subject_id": subject_id, "session_id": session_id, "teacher_id": teacher_id, "student1_id": student1_id, "student2_id": student2_id})

    status, body = req(
        "POST",
        "/api/v1/question-sets",
        admin,
        {
            "subject_id": subject_id,
            "title": f"SET-S {run_id}",
            "jenjang": "SMP",
            "level_id": level_id,
        },
    )
    expect(checks, "setup_create_question_set", status, {201}, body)
    qset_id = (body.get("data") or {}).get("id")
    if not qset_id:
        fatal("create question set", status, body)

    status, body = req(
        "POST",
        f"/api/v1/question-sets/{qset_id}/questions",
        admin,
        {
            "type": "mc_single",
            "stem": "2 + 2 = ?",
            "order_no": 1,
            "weight": 1,
            "options": [
                {"label": "A", "content": "4", "is_correct": True},
                {"label": "B", "content": "5", "is_correct": False},
            ],
        },
    )
    expect(checks, "setup_create_question", status, {201}, body)
    question_id = (body.get("data") or {}).get("id")
    if not question_id:
        fatal("create question", status, body)

    now = datetime.datetime.now(datetime.timezone.utc)
    active_start = now - datetime.timedelta(minutes=5)
    active_end = now + datetime.timedelta(hours=1)
    past_start = now - datetime.timedelta(hours=3)
    past_end = now - datetime.timedelta(hours=2)

    status, body = req(
        "POST",
        "/api/v1/exams",
        admin,
        {
            "subject_id": subject_id,
            "teacher_id": teacher_id,
            "session_id": session_id,
            "title": f"EXAM-S-ACTIVE {run_id}",
            "starts_at": active_start.strftime("%Y-%m-%dT%H:%M:%SZ"),
            "ends_at": active_end.strftime("%Y-%m-%dT%H:%M:%SZ"),
            "duration_minutes": 60,
            "shuffle_questions": False,
            "shuffle_options": False,
            "scoring_mode": "partial",
        },
    )
    expect(checks, "setup_create_active_exam", status, {201}, body)
    active_exam_id = (body.get("data") or {}).get("id")
    if not active_exam_id:
        fatal("create active exam", status, body)

    status, body = req(
        "PUT",
        f"/api/v1/exams/{active_exam_id}/question-sets",
        admin,
        {"items": [{"question_set_id": qset_id}]},
    )
    expect(checks, "setup_attach_qset_active_exam", status, {200}, body)

    status, body = req(
        "PUT",
        f"/api/v1/exams/{active_exam_id}/targets",
        admin,
        {"level_ids": [], "group_ids": [], "student_ids": [student1_id, student2_id]},
    )
    expect(checks, "setup_attach_targets_active_exam", status, {200}, body)

    status, body = req(
        "POST",
        f"/api/v1/exams/{active_exam_id}/tokens",
        admin,
        {"length": 6},
    )
    expect(checks, "setup_create_exam_token", status, {201}, body)
    active_exam_token = ((body.get("data") or {}).get("token") or "").strip()
    if not active_exam_token:
        fatal("setup create token missing value", status, body)

    status, body = req(
        "PATCH",
        f"/api/v1/exams/{active_exam_id}",
        admin,
        {"status": "published"},
    )
    expect(checks, "setup_publish_active_exam", status, {200}, body)

    status, body = req(
        "POST",
        "/api/v1/exams",
        admin,
        {
            "subject_id": subject_id,
            "teacher_id": teacher_id,
            "session_id": session_id,
            "title": f"EXAM-S-ENDED {run_id}",
            "starts_at": past_start.strftime("%Y-%m-%dT%H:%M:%SZ"),
            "ends_at": past_end.strftime("%Y-%m-%dT%H:%M:%SZ"),
            "duration_minutes": 60,
            "shuffle_questions": False,
            "shuffle_options": False,
            "scoring_mode": "partial",
        },
    )
    expect(checks, "setup_create_ended_exam", status, {201}, body)
    ended_exam_id = (body.get("data") or {}).get("id")
    if not ended_exam_id:
        fatal("create ended exam", status, body)

    status, body = req(
        "PUT",
        f"/api/v1/exams/{ended_exam_id}/targets",
        admin,
        {"level_ids": [], "group_ids": [], "student_ids": [student1_id]},
    )
    expect(checks, "setup_attach_targets_ended_exam", status, {200}, body)

    status, body = req(
        "PATCH",
        f"/api/v1/exams/{ended_exam_id}",
        admin,
        {"status": "published"},
    )
    expect(checks, "setup_publish_ended_exam", status, {200}, body)

    # Auth checks
    status, body = req("GET", "/api/v1/student/exams")
    expect(checks, "student_list_exams_without_auth", status, {401}, body)

    status, login1 = req(
        "POST",
        "/api/v1/auth/login",
        body={"username": student1_username, "password": student_password},
    )
    expect(checks, "student1_login", status, {200}, login1)
    student1_token = (login1.get("data") or {}).get("access_token", "")

    status, login2 = req(
        "POST",
        "/api/v1/auth/login",
        body={"username": student2_username, "password": student_password},
    )
    expect(checks, "student2_login", status, {200}, login2)
    student2_token = (login2.get("data") or {}).get("access_token", "")

    if not student1_token or not student2_token:
        fatal("student login token empty", 401, {"student1_token": bool(student1_token), "student2_token": bool(student2_token)})

    status, body = req("GET", "/api/v1/me", student1_token)
    expect(checks, "student1_me", status, {200}, body)

    status, body = req("GET", "/api/v1/admin/programs", student1_token)
    expect(checks, "student1_admin_forbidden", status, {403}, body)

    status, body = req("GET", "/api/v1/student/exams?limit=20&offset=0", student1_token)
    expect(checks, "student1_list_exams", status, {200}, body)

    status, body = req("DELETE", f"/api/v1/student/exams/{active_exam_id}/dismiss", student1_token)
    expect(checks, "student1_dismiss_active_exam_blocked", status, {409}, body)

    status, body = req("DELETE", f"/api/v1/student/exams/{ended_exam_id}/dismiss", student1_token)
    expect(checks, "student1_dismiss_ended_exam", status, {200}, body)

    status, body = req("POST", f"/api/v1/student/exams/{active_exam_id}/join", student1_token, {})
    expect(checks, "student1_join_without_token_required", status, {400}, body)

    status, body = req(
        "POST",
        f"/api/v1/student/exams/{active_exam_id}/join",
        student1_token,
        {"token": "000000"},
    )
    expect(checks, "student1_join_with_invalid_token", status, {400}, body)

    status, body = req(
        "POST",
        f"/api/v1/student/exams/{active_exam_id}/join",
        student1_token,
        {"token": active_exam_token},
    )
    expect(checks, "student1_join_with_valid_token", status, {200}, body)
    session1_id = ((body.get("data") or {}).get("session") or {}).get("id") or (body.get("data") or {}).get("session_id")
    if not session1_id:
        fatal("student1 join missing session_id", status, body)

    status, body = req("GET", f"/api/v1/student/exams/{active_exam_id}/session", student1_token)
    expect(checks, "student1_get_active_session_by_exam", status, {200}, body)

    status, body = req("GET", f"/api/v1/student/sessions/{session1_id}", student1_token)
    expect(checks, "student1_get_session", status, {200}, body)

    status, body = req("GET", f"/api/v1/student/sessions/{session1_id}/questions", student1_token)
    expect(checks, "student1_get_questions", status, {200}, body)
    questions = ((body.get("data") or {}).get("questions") or [])
    if not questions:
        fatal("student1 questions empty", status, body)
    qid = questions[0].get("id")
    if not qid:
        fatal("student1 first question id missing", status, questions[0])

    q0_raw = json.dumps(questions[0])
    no_sensitive = ("correct_answer" not in q0_raw) and ("answer_key" not in q0_raw)
    checks.append(("student1_questions_hide_answer_key", 200 if no_sensitive else 500, [200], no_sensitive, None, "ensure no correct_answer/answer_key"))

    status, body = req("GET", f"/api/v1/student/sessions/{session1_id}/answers", student1_token)
    expect(checks, "student1_get_answers_initial", status, {200}, body)

    status, body = req(
        "POST",
        f"/api/v1/student/sessions/{session1_id}/answers",
        student1_token,
        {"question_id": qid},
    )
    expect(checks, "student1_save_answer_invalid_payload", status, {400}, body)

    status, body = req(
        "POST",
        f"/api/v1/student/sessions/{session1_id}/answers",
        student1_token,
        {"question_id": qid, "answer_json": {"choice": "A"}},
    )
    expect(checks, "student1_save_answer_valid", status, {200}, body)

    status, body = req(
        "POST",
        f"/api/v1/student/sessions/{session1_id}/answers",
        student1_token,
        {"question_id": qid, "answer_json": {"choice": "B"}},
    )
    expect(checks, "student1_save_answer_upsert_same_question", status, {200}, body)

    status, body = req("GET", f"/api/v1/student/sessions/{session1_id}/answers", student1_token)
    expect(checks, "student1_get_answers_after_upsert", status, {200}, body)
    answers = body.get("data") or []
    same_q_answers = [a for a in answers if a.get("question_id") == qid]
    upsert_ok = len(same_q_answers) == 1
    checks.append(("student1_answers_not_duplicated", 200 if upsert_ok else 500, [200], upsert_ok, None, "same question should stay single row"))

    status, body = req(
        "POST",
        f"/api/v1/student/sessions/{session1_id}/heartbeat",
        student1_token,
        {"tab": "active"},
    )
    expect(checks, "student1_heartbeat", status, {200}, body)

    status, body = req("GET", f"/api/v1/exams/{active_exam_id}/questions", student1_token)
    expect(checks, "student1_get_questions_compat", status, {200}, body)

    status, body = req(
        "POST",
        f"/api/v1/sessions/{session1_id}/answers",
        student1_token,
        {"question_id": qid, "answer_json": {"choice": "A"}},
    )
    expect(checks, "student1_save_answer_compat", status, {200}, body)

    status, body = req("POST", f"/api/v1/student/sessions/{session1_id}/submit", student1_token, {})
    expect(checks, "student1_submit", status, {200}, body)

    status, body = req("POST", f"/api/v1/student/sessions/{session1_id}/submit", student1_token, {})
    expect(checks, "student1_submit_twice_conflict", status, {409}, body)

    status, body = req("POST", f"/api/v1/sessions/{session1_id}/finish", student1_token, {})
    expect(checks, "student1_finish_compat_after_submit_conflict", status, {409}, body)

    status, body = req("DELETE", f"/api/v1/student/exams/{active_exam_id}/dismiss", student1_token)
    expect(checks, "student1_dismiss_active_exam_after_finish", status, {200}, body)

    status, body = req("GET", "/api/v1/student/results?limit=20&offset=0", student1_token)
    expect(checks, "student1_results", status, {200}, body)

    status, body = req("GET", "/api/v1/student/announcements?limit=20&offset=0", student1_token)
    expect(checks, "student1_announcements", status, {200}, body)

    status, body = req("GET", "/api/v1/student/attendance/history?limit=20&offset=0", student1_token)
    expect(checks, "student1_attendance_history", status, {200}, body)

    # student2 ownership checks
    status, body = req(
        "POST",
        f"/api/v1/student/exams/{active_exam_id}/join",
        student2_token,
        {"token": active_exam_token},
    )
    expect(checks, "student2_join_with_valid_token", status, {200}, body)
    session2_id = ((body.get("data") or {}).get("session") or {}).get("id") or (body.get("data") or {}).get("session_id")
    if not session2_id:
        fatal("student2 join missing session_id", status, body)

    status, body = req("GET", f"/api/v1/student/sessions/{session1_id}", student2_token)
    expect(checks, "student2_get_session_other_user_forbidden", status, {403}, body)

    status, body = req("GET", f"/api/v1/student/sessions/{session1_id}/questions", student2_token)
    expect(checks, "student2_get_questions_other_user_forbidden", status, {403}, body)

    status, body = req("GET", f"/api/v1/student/sessions/{session1_id}/answers", student2_token)
    expect(checks, "student2_get_answers_other_user_forbidden", status, {403}, body)

    status, body = req(
        "POST",
        f"/api/v1/student/sessions/{session1_id}/answers",
        student2_token,
        {"question_id": qid, "answer_json": {"choice": "A"}},
    )
    expect(checks, "student2_save_answer_other_user_not_found", status, {404}, body)

    status, body = req("POST", f"/api/v1/student/sessions/{session1_id}/submit", student2_token, {})
    expect(checks, "student2_submit_other_user_not_found", status, {404}, body)

    status, body = req(
        "POST",
        f"/api/v1/sessions/{session1_id}/answers",
        student2_token,
        {"question_id": qid, "answer_json": {"choice": "A"}},
    )
    expect(checks, "student2_save_answer_compat_other_user_forbidden", status, {403}, body)

    status, body = req("POST", f"/api/v1/sessions/{session1_id}/finish", student2_token, {})
    expect(checks, "student2_finish_compat_other_user_not_found", status, {404}, body)

    status, body = req("POST", f"/api/v1/student/sessions/{session2_id}/verify-token", student2_token, {"token": "bad"})
    expect(checks, "student2_verify_token_invalid", status, {400}, body)

    failed = [item for item in checks if not item[3]]
    for name, status, allowed, ok, err, note in checks:
        print(f"{'PASS' if ok else 'FAIL'} {name}: status={status} expected={allowed}")
        if note:
            print(f"  note={note}")
        if err:
            print(f"  error={err}")

    print(
        json.dumps(
            {
                "run_id": run_id,
                "active_exam_id": active_exam_id,
                "ended_exam_id": ended_exam_id,
                "session1_id": session1_id,
                "session2_id": session2_id,
                "question_id": question_id,
                "total_checks": len(checks),
                "failed_checks": len(failed),
            }
        )
    )
    if failed:
        sys.exit(2)


if __name__ == "__main__":
    main()
