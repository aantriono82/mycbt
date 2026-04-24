import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  stages: [
    { duration: '1m', target: 50 },
    { duration: '3m', target: 100 },
    { duration: '1m', target: 0 },
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'],
    http_req_failed: ['rate<0.01'],
  },
};

function requiredEnv(name) {
  const v = __ENV[name];
  if (!v || String(v).trim() === '') {
    throw new Error(`missing required env: ${name}`);
  }
  return String(v).trim();
}

function resolveToken(baseURL) {
  if (__ENV.TOKEN && String(__ENV.TOKEN).trim() !== '') {
    return String(__ENV.TOKEN).trim();
  }
  const username = __ENV.USERNAME || 'siswa1';
  const password = requiredEnv('PASSWORD');
  const res = http.post(
    `${baseURL}/api/v1/auth/login`,
    JSON.stringify({ username, password }),
    { headers: { 'Content-Type': 'application/json' }, tags: { name: 'auth_login' } },
  );
  check(res, { 'login 200': (r) => r.status === 200 });
  return res.json('data.access_token');
}

export function setup() {
  const baseURL = requiredEnv('BASE_URL');
  const examID = requiredEnv('EXAM_ID');
  const examToken = requiredEnv('EXAM_TOKEN');
  const questionID = requiredEnv('QUESTION_ID');
  const token = resolveToken(baseURL);
  const headers = {
    Authorization: `Bearer ${token}`,
    'Content-Type': 'application/json',
  };
  const join = http.post(
    `${baseURL}/api/v1/student/exams/${examID}/join`,
    JSON.stringify({ token: examToken }),
    { headers, tags: { name: 'student_join_setup' } },
  );
  check(join, { 'setup join 200': (r) => r.status === 200 });
  const sessionID = join.json('data.session_id');
  if (!sessionID) {
    throw new Error('setup failed: missing session_id');
  }
  return { baseURL, examID, examToken, questionID, token, sessionID };
}

export default function (cfg) {
  const headers = {
    Authorization: `Bearer ${cfg.token}`,
    'Content-Type': 'application/json',
  };

  let sessionID = cfg.sessionID;
  if (__ENV.JOIN_EACH_ITERATION === '1') {
    const join = http.post(
      `${cfg.baseURL}/api/v1/student/exams/${cfg.examID}/join`,
      JSON.stringify({ token: cfg.examToken }),
      { headers, tags: { name: 'student_join' } },
    );
    const joinOK = check(join, { 'join 200': (r) => r.status === 200 });
    if (!joinOK) {
      sleep(1);
      return;
    }
    sessionID = join.json('data.session_id');
  }

  if (!sessionID) {
    check(null, { 'session_id exists': () => false });
    sleep(1);
    return;
  }

  const answer = http.post(
    `${cfg.baseURL}/api/v1/student/sessions/${sessionID}/answers`,
    JSON.stringify({
      question_id: cfg.questionID,
      answer_json: { value: 'k6-answer' },
    }),
    { headers, tags: { name: 'session_answer' } },
  );
  check(answer, { 'answer 200': (r) => r.status === 200 });

  sleep(1);
}
