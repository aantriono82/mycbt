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


def expect(checks, name, status, allowed, body=None):
    ok = status in allowed
    err = body.get("error") if isinstance(body, dict) else None
    checks.append((name, status, sorted(allowed), ok, err))


def require_ok(status, body, message):
    if status not in (200, 201):
        print(f"FATAL {message}: status={status} body={body}")
        sys.exit(1)


def main():
    checks = []
    run_id = str(int(time.time()))

    status, login = req(
        "POST",
        "/api/v1/auth/login",
        body={"username": "admin", "password": "admin12345"},
    )
    require_ok(status, login, "admin login")
    admin = login["data"]["access_token"]

    _, program = req(
        "POST",
        "/api/v1/admin/programs",
        admin,
        {"code": f"PRG-T-{run_id}", "name": f"Program T {run_id}"},
    )
    program_id = program["data"]["id"]
    _, level_1 = req(
        "POST",
        "/api/v1/admin/levels",
        admin,
        {"name": f"LevelT1 {run_id}", "kelas": 10},
    )
    _, level_2 = req(
        "POST",
        "/api/v1/admin/levels",
        admin,
        {"name": f"LevelT2 {run_id}", "kelas": 11},
    )
    _, group_1 = req(
        "POST",
        "/api/v1/admin/groups",
        admin,
        {"name": f"GroupT1 {run_id}"},
    )
    _, group_2 = req(
        "POST",
        "/api/v1/admin/groups",
        admin,
        {"name": f"GroupT2 {run_id}"},
    )

    level1_id = level_1["data"]["id"]
    level2_id = level_2["data"]["id"]
    group1_id = group_1["data"]["id"]
    group2_id = group_2["data"]["id"]

    subject_code_a = f"SBJA{run_id}"
    subject_code_b = f"SBJB{run_id}"
    _, subject_a = req(
        "POST",
        "/api/v1/admin/subjects",
        admin,
        {"code": subject_code_a, "name": f"Subject A {run_id}"},
    )
    _, subject_b = req(
        "POST",
        "/api/v1/admin/subjects",
        admin,
        {"code": subject_code_b, "name": f"Subject B {run_id}"},
    )
    subject_a_id = subject_a["data"]["id"]
    subject_b_id = subject_b["data"]["id"]

    _, session = req(
        "POST",
        "/api/v1/admin/sessions",
        admin,
        {"name": f"Sesi T {run_id}", "start_time": "00:00", "end_time": "23:59"},
    )
    session_id = session["data"]["id"]

    teacher_1_username = f"guruA_{run_id}"
    teacher_2_username = f"guruB_{run_id}"
    teacher_password = "Guru12345!"

    _, teacher_1 = req(
        "POST",
        "/api/v1/admin/teachers",
        admin,
        {
            "username": teacher_1_username,
            "password": teacher_password,
            "name": f"Guru A {run_id}",
            "email": f"{teacher_1_username}@x.test",
            "nip": f"{run_id}1",
            "jenjang": "SMA",
            "mapel_codes": subject_code_a,
            "group_names": f"GroupT1 {run_id}",
            "level_names": f"LevelT1 {run_id}",
        },
    )
    _, teacher_2 = req(
        "POST",
        "/api/v1/admin/teachers",
        admin,
        {
            "username": teacher_2_username,
            "password": teacher_password,
            "name": f"Guru B {run_id}",
            "email": f"{teacher_2_username}@x.test",
            "nip": f"{run_id}2",
            "jenjang": "SMA",
            "mapel_codes": subject_code_b,
            "group_names": f"GroupT2 {run_id}",
            "level_names": f"LevelT2 {run_id}",
        },
    )
    teacher_1_id = teacher_1["data"]["id"]
    teacher_2_id = teacher_2["data"]["id"]

    student_username = f"siswaT_{run_id}"
    _, student = req(
        "POST",
        "/api/v1/admin/students",
        admin,
        {
            "username": student_username,
            "password": "Siswa12345!",
            "name": f"Siswa T {run_id}",
            "email": f"{student_username}@x.test",
            "nis": run_id,
            "jenjang": "SMA",
            "program_id": program_id,
            "level_id": level1_id,
            "group_id": group1_id,
        },
    )
    student_id = student["data"]["id"]

    status, login_1 = req(
        "POST",
        "/api/v1/auth/login",
        body={"username": teacher_1_username, "password": teacher_password},
    )
    expect(checks, "teacher1_login", status, {200}, login_1)
    teacher_1_token = login_1.get("data", {}).get("access_token", "")

    status, login_2 = req(
        "POST",
        "/api/v1/auth/login",
        body={"username": teacher_2_username, "password": teacher_password},
    )
    expect(checks, "teacher2_login", status, {200}, login_2)
    teacher_2_token = login_2.get("data", {}).get("access_token", "")

    for name, path in [
        ("teacher1_me", "/api/v1/me"),
        ("teacher1_my_assignments", "/api/v1/lookups/my-assignments"),
        ("teacher1_qsets_list", "/api/v1/question-sets?limit=10&offset=0"),
        ("teacher1_exams_list", "/api/v1/exams?limit=10&offset=0"),
    ]:
        status, body = req("GET", path, teacher_1_token)
        expect(checks, name, status, {200}, body)

    status, body = req("GET", "/api/v1/admin/programs", teacher_1_token)
    expect(checks, "teacher1_admin_programs_forbidden", status, {403}, body)

    status, body = req(
        "POST",
        "/api/v1/question-sets",
        teacher_1_token,
        {
            "subject_id": subject_a_id,
            "title": f"SET-A {run_id}",
            "jenjang": "SMA",
            "level_id": level1_id,
        },
    )
    expect(checks, "teacher1_create_set_assigned", status, {201}, body)
    set_1_id = (body.get("data") or {}).get("id")

    status, body = req(
        "POST",
        "/api/v1/question-sets",
        teacher_1_token,
        {
            "subject_id": subject_b_id,
            "title": f"SET-BLOCK {run_id}",
            "jenjang": "SMA",
            "level_id": level1_id,
        },
    )
    expect(checks, "teacher1_create_set_unassigned_forbidden", status, {403}, body)

    status, body = req("GET", f"/api/v1/question-sets/{set_1_id}", teacher_1_token)
    expect(checks, "teacher1_get_own_set", status, {200}, body)
    status, body = req("GET", f"/api/v1/question-sets/{set_1_id}", teacher_2_token)
    expect(checks, "teacher2_get_other_set_forbidden", status, {403}, body)

    status, body = req(
        "POST",
        f"/api/v1/question-sets/{set_1_id}/questions",
        teacher_1_token,
        {
            "type": "mc_single",
            "stem": "2+2=?",
            "order_no": 1,
            "options": [
                {"label": "A", "content": "4", "is_correct": True},
                {"label": "B", "content": "5", "is_correct": False},
            ],
        },
    )
    expect(checks, "teacher1_create_question", status, {201}, body)
    status, body = req("GET", f"/api/v1/question-sets/{set_1_id}/questions", teacher_1_token)
    expect(checks, "teacher1_list_questions", status, {200}, body)

    now = datetime.datetime.now(datetime.timezone.utc)
    ends_at = now + datetime.timedelta(hours=2)
    status, body = req(
        "POST",
        "/api/v1/exams",
        teacher_1_token,
        {
            "subject_id": subject_a_id,
            "title": f"EXAM-A {run_id}",
            "session_id": session_id,
            "starts_at": now.strftime("%Y-%m-%dT%H:%M:%SZ"),
            "ends_at": ends_at.strftime("%Y-%m-%dT%H:%M:%SZ"),
            "duration_minutes": 60,
            "shuffle_questions": False,
            "shuffle_options": False,
            "scoring_mode": "partial",
        },
    )
    expect(checks, "teacher1_create_exam_assigned_subject", status, {201}, body)
    exam_1_id = (body.get("data") or {}).get("id")

    status, body = req(
        "POST",
        "/api/v1/exams",
        teacher_1_token,
        {
            "subject_id": subject_b_id,
            "title": f"EXAM-BLOCK {run_id}",
            "session_id": session_id,
            "starts_at": now.strftime("%Y-%m-%dT%H:%M:%SZ"),
            "ends_at": ends_at.strftime("%Y-%m-%dT%H:%M:%SZ"),
            "duration_minutes": 60,
            "shuffle_questions": False,
            "shuffle_options": False,
            "scoring_mode": "partial",
        },
    )
    expect(checks, "teacher1_create_exam_unassigned_subject_forbidden", status, {403}, body)

    status, body = req("GET", f"/api/v1/exams/{exam_1_id}", teacher_1_token)
    expect(checks, "teacher1_get_own_exam", status, {200}, body)
    status, body = req("GET", f"/api/v1/exams/{exam_1_id}", teacher_2_token)
    expect(checks, "teacher2_get_other_exam_forbidden", status, {403}, body)

    status, body = req(
        "PUT",
        f"/api/v1/exams/{exam_1_id}/question-sets",
        teacher_1_token,
        {"items": [{"question_set_id": set_1_id}]},
    )
    expect(checks, "teacher1_put_question_sets", status, {200}, body)

    status, body = req(
        "PUT",
        f"/api/v1/exams/{exam_1_id}/targets",
        teacher_1_token,
        {"level_ids": [level1_id], "group_ids": [group1_id], "student_ids": [student_id]},
    )
    expect(checks, "teacher1_put_targets_assigned", status, {200}, body)

    status, body = req(
        "PUT",
        f"/api/v1/exams/{exam_1_id}/targets",
        teacher_1_token,
        {"level_ids": [level2_id], "group_ids": [], "student_ids": []},
    )
    expect(checks, "teacher1_put_targets_unassigned_level_forbidden", status, {403}, body)

    status, body = req(
        "PUT",
        f"/api/v1/exams/{exam_1_id}/targets",
        teacher_1_token,
        {"level_ids": [], "group_ids": [group2_id], "student_ids": []},
    )
    expect(checks, "teacher1_put_targets_unassigned_group_forbidden", status, {403}, body)

    status, body = req("GET", f"/api/v1/exams/{exam_1_id}/tokens", teacher_1_token)
    expect(checks, "teacher1_list_tokens", status, {200}, body)

    status, body = req(
        "POST",
        f"/api/v1/exams/{exam_1_id}/tokens",
        teacher_1_token,
        {"token": "654321", "is_active": True},
    )
    expect(checks, "teacher1_create_token", status, {201}, body)

    status, body = req("POST", f"/api/v1/exams/{exam_1_id}/tokens/rotate", teacher_2_token, {})
    expect(checks, "teacher2_rotate_other_exam_token_forbidden", status, {403}, body)

    status, body = req("PATCH", f"/api/v1/exams/{exam_1_id}", teacher_1_token, {"status": "published"})
    expect(checks, "teacher1_publish_exam", status, {200}, body)

    status, body = req("GET", f"/api/v1/exams/{exam_1_id}/results?limit=10&offset=0", teacher_1_token)
    expect(checks, "teacher1_results", status, {200}, body)
    status, body = req("GET", f"/api/v1/exams/{exam_1_id}/results?limit=10&offset=0", teacher_2_token)
    expect(checks, "teacher2_results_other_exam_forbidden", status, {403}, body)

    status, body = req("GET", f"/api/v1/exams/{exam_1_id}/monitor/sessions?limit=10&offset=0", teacher_1_token)
    expect(checks, "teacher1_monitor_sessions", status, {200}, body)
    status, body = req("GET", f"/api/v1/exams/{exam_1_id}/monitor/sessions?limit=10&offset=0", teacher_2_token)
    expect(checks, "teacher2_monitor_other_exam_forbidden", status, {403}, body)

    failed = [item for item in checks if not item[3]]
    for name, status, allowed, ok, err in checks:
        print(f"{'PASS' if ok else 'FAIL'} {name}: status={status} expected={allowed}")
        if err:
            print(f"  error={err}")

    print(
        json.dumps(
            {
                "run_id": run_id,
                "teacher_1_id": teacher_1_id,
                "teacher_2_id": teacher_2_id,
                "exam_id": exam_1_id,
                "total_checks": len(checks),
                "failed_checks": len(failed),
            }
        )
    )
    if failed:
        sys.exit(2)


if __name__ == "__main__":
    main()
