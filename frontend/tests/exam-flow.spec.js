import { expect, test } from '@playwright/test'

test.setTimeout(150000)

const EXAM_ID = 'exam-1'
const SESSION_ID = 'session-1'

const examListPayload = [
  {
    id: EXAM_ID,
    title: 'Ujian Matematika',
    subject_name: 'Matematika',
    teacher_name: 'Bu Rina',
    starts_at: '2026-01-01T00:00:00Z',
    ends_at: '2099-12-31T23:59:59Z',
    can_join: true,
    duration_minutes: 60,
    active_token: 'TOK123',
    session_status: '',
  },
]

const questionsPayload = {
  questions: [
    {
      id: 'q-1',
      type: 'mc_single',
      stem: '<p>2 + 2 = ?</p>',
      options: [
        { id: 'opt-a', content: '<p>4</p>' },
        { id: 'opt-b', content: '<p>5</p>' },
      ],
    },
    {
      id: 'q-2',
      type: 'mc_single',
      stem: '<p>3 + 3 = ?</p>',
      options: [
        { id: 'opt-c', content: '<p>6</p>' },
        { id: 'opt-d', content: '<p>7</p>' },
      ],
    },
    {
      id: 'q-3',
      type: 'mc_single',
      stem: '<p>10 - 4 = ?</p>',
      options: [
        { id: 'opt-e', content: '<p>6</p>' },
        { id: 'opt-f', content: '<p>5</p>' },
      ],
    },
  ],
}

const resultsPayload = [
  {
    session_id: SESSION_ID,
    exam_title: 'Ujian Matematika',
    subject: 'Matematika',
    submitted_at: '2026-04-24T10:15:00+07:00',
    score: 100,
    correct_count: 3,
    auto_scorable_questions: 3,
    total_questions: 3,
    session_status: 'submitted',
  },
]

const buildAnswerRowsFromState = (answersByQuestion) =>
  Object.entries(answersByQuestion).map(([questionId, answerJson]) => ({
    id: `ans-${questionId}`,
    question_id: questionId,
    answer_json: answerJson,
  }))

async function installApiMocks(page, options = {}) {
  const { remainingSeconds = 65, resultsData = resultsPayload } = options
  const state = {
    saveAnswerCalls: 0,
    submitCalls: 0,
    heartbeatCalls: 0,
    answersByQuestion: {},
  }

  await page.route('**/api/**', async (route) => {
    const req = route.request()
    const url = new URL(req.url())
    const path = url.pathname
    const method = req.method()

    if (path === '/api/v1/auth/login' && method === 'POST') {
      return route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          data: { token: 'e2e-token' },
        }),
      })
    }

    if (path === '/api/v1/me' && method === 'GET') {
      return route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          data: { id: 1, name: 'Siswa E2E', username: 'siswa1', role: 'student' },
        }),
      })
    }

    if (path === '/healthz' && method === 'GET') {
      return route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ ok: true, time: '2026-04-24T10:00:00+07:00' }),
      })
    }

    if (path === '/api/v1/student/notifications/stream' && method === 'GET') {
      return route.fulfill({
        status: 200,
        contentType: 'text/event-stream',
        body: '',
      })
    }

    if (path === '/api/v1/student/announcements' && method === 'GET') {
      return route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ data: [] }),
      })
    }

    if (path === '/api/v1/student/exams' && method === 'GET') {
      return route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ data: examListPayload }),
      })
    }

    if (path === '/api/v1/student/results' && method === 'GET') {
      return route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ data: resultsData }),
      })
    }

    if (path === `/api/v1/student/exams/${EXAM_ID}/join` && method === 'POST') {
      return route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ data: { session_id: SESSION_ID } }),
      })
    }

    if (path === `/api/v1/student/sessions/${SESSION_ID}/verify-token` && method === 'POST') {
      return route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ data: { valid: true } }),
      })
    }

    if (path === `/api/v1/student/sessions/${SESSION_ID}` && method === 'GET') {
      return route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          data: {
            id: SESSION_ID,
            remaining_seconds: remainingSeconds,
            exam: { id: EXAM_ID, title: 'Ujian Matematika' },
          },
        }),
      })
    }

    if (path === `/api/v1/student/sessions/${SESSION_ID}/questions` && method === 'GET') {
      return route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ data: questionsPayload }),
      })
    }

    if (path === `/api/v1/student/sessions/${SESSION_ID}/answers` && method === 'GET') {
      return route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ data: buildAnswerRowsFromState(state.answersByQuestion) }),
      })
    }

    if (path === `/api/v1/student/sessions/${SESSION_ID}/answers` && method === 'POST') {
      const payload = req.postDataJSON() || {}
      if (payload.question_id) {
        state.answersByQuestion[payload.question_id] = payload.answer_json
      }
      state.saveAnswerCalls += 1
      return route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ data: { saved: true } }),
      })
    }

    if (path === `/api/v1/student/sessions/${SESSION_ID}/submit` && method === 'POST') {
      state.submitCalls += 1
      return route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ data: { submitted: true } }),
      })
    }

    if (path === `/api/v1/student/sessions/${SESSION_ID}/heartbeat` && method === 'POST') {
      state.heartbeatCalls += 1
      return route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ data: { ok: true } }),
      })
    }

    return route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ data: [] }),
    })
  })

  return state
}

async function loginAsStudent(page) {
  await page.goto('/#/login')
  await expect(page.getByRole('heading', { name: 'Login Atiga CBT' })).toBeVisible()
  await page.getByPlaceholder('Masukkan email atau username').fill('siswa1')
  await page.getByPlaceholder('Masukkan password').fill('siswa123')
  await Promise.all([
    page.waitForResponse(
      (resp) => resp.url().includes('/api/v1/auth/login') && resp.request().method() === 'POST',
    ),
    page.getByRole('button', { name: 'Masuk', exact: true }).click(),
  ])
  await expect(page).toHaveURL(/\/#\/student\/dashboard/)
  await expect(page.getByText('Dashboard Siswa')).toBeVisible()
}

async function goToWorkspaceAndStartExam(page) {
  await page.goto('/#/student/ujian')
  await expect(page.getByRole('heading', { name: 'Ruang Ujian' })).toBeVisible()
  await page.getByRole('button', { name: 'Masuk Ruang Ujian' }).first().click()
  await expect(page).toHaveURL(/\/#\/student\/ujian\/exam-1\/token/)
  await expect(page.getByRole('heading', { name: 'Verifikasi Token Ujian' })).toBeVisible()
  await page.getByPlaceholder('Masukkan token dari pengawas').fill('TOK123')
  const tokenCard = page.locator('.mx-auto.w-full.max-w-2xl')
  await Promise.all([
    page.waitForResponse(
      (resp) =>
        resp.url().includes(`/api/v1/student/exams/${EXAM_ID}/join`) &&
        resp.request().method() === 'POST',
    ),
    tokenCard.getByRole('button', { name: 'Masuk Ruang Ujian' }).click(),
  ])
  await expect(page).toHaveURL(/\/#\/student\/workspace\/session-1/)
  const firstQuestionMarker = page.getByText('SOAL NOMOR: 1')
  await expect(firstQuestionMarker).toBeVisible({ timeout: 10000 })
}

async function selectMcOption(page, optionIndex) {
  await page.locator('main label').nth(optionIndex).click()
}

test.beforeEach(async ({ page }) => {
  await page.addInitScript(() => {
    if (!sessionStorage.getItem('__e2e_bootstrapped__')) {
      localStorage.clear()
      localStorage.removeItem('atigacbt_session_token_ok_session-1')
      localStorage.setItem(
        'atigacbt:exam-store',
        JSON.stringify({
          sessionId: 'session-1',
          startTime: '2026-04-24T10:00:00+07:00',
          questions: [],
          answers: {},
          currentQuestionIdx: 0,
          examTitle: 'AtigaCBT Workspace',
          timeLeft: 1800,
        }),
      )
      sessionStorage.setItem('__e2e_bootstrapped__', '1')
    }
    class MockEventSource {
      constructor(url) {
        this.url = url
      }
      addEventListener() {}
      close() {}
    }
    window.EventSource = MockEventSource
  })
})

test('exam flow: login, mulai ujian, jawab, submit, dan lihat hasil', async ({ page }) => {
  const state = await installApiMocks(page)

  // 1) LOGIN
  await loginAsStudent(page)

  // 2) MULAI UJIAN
  await goToWorkspaceAndStartExam(page)
  const timer = page.locator('header .font-mono').first()
  const initialTimer = await timer.textContent()
  await page.waitForTimeout(1300)
  await expect
    .poll(async () => {
      const next = await timer.textContent()
      return next !== initialTimer
    })
    .toBeTruthy()

  // 3) JAWAB SOAL
  await selectMcOption(page, 0)
  await expect(page.locator('input[name="mc-q-1"]').nth(0)).toBeChecked()
  await page.getByRole('button', { name: 'Soal Berikutnya' }).click()
  await expect(page.getByText('SOAL NOMOR: 2')).toBeVisible()
  await page.locator('[data-qnav-idx="2"]').click()
  await expect(page.getByText('SOAL NOMOR: 3')).toBeVisible()
  await page.locator('[data-qnav-idx="0"]').click()
  await expect(page.getByText('SOAL NOMOR: 1')).toBeVisible()
  await selectMcOption(page, 1)
  await expect(page.locator('input[name="mc-q-1"]').nth(1)).toBeChecked()
  await expect(page.locator('input[name="mc-q-1"]').nth(0)).not.toBeChecked()

  // 4) SUBMIT
  await page.locator('[data-qnav-idx="2"]').click()
  await page.getByRole('button', { name: 'Selesai', exact: true }).click()
  await expect(page.getByText('Konfirmasi Selesai')).toBeVisible()
  await Promise.all([
    page.waitForResponse(
      (resp) =>
        resp.url().includes(`/api/v1/student/sessions/${SESSION_ID}/submit`) &&
        resp.request().method() === 'POST',
    ),
    page.getByRole('button', { name: 'YA, SAYA YAKIN' }).click(),
  ])
  await expect(page).toHaveURL(/\/#\/student\/hasil/, { timeout: 15000 })

  // 5) HASIL
  await expect(page.getByRole('heading', { name: 'Hasil Ujian' })).toBeVisible()
  await expect(page.getByText('100').first()).toBeVisible()
  await expect(page.getByText('Benar: 3/3')).toBeVisible()
  await expect(page.getByText('Submitted')).toBeVisible()

  expect(state.saveAnswerCalls).toBeGreaterThan(0)
  expect(state.submitCalls).toBe(1)
})

test('edge case: waktu ujian habis memicu auto-submit', async ({ page }) => {
  const state = await installApiMocks(page, { remainingSeconds: 60 })

  await loginAsStudent(page)
  await goToWorkspaceAndStartExam(page)
  await expect(page.locator('header .font-mono').first()).toContainText(/00:0[01]:\d{2}/)
  await page.waitForResponse(
    (resp) =>
      resp.url().includes(`/api/v1/student/sessions/${SESSION_ID}/submit`) &&
      resp.request().method() === 'POST',
    { timeout: 80000 },
  )
  await expect(page.getByText('Ujian Selesai!')).toBeVisible({ timeout: 20000 })
  expect(state.submitCalls).toBeGreaterThanOrEqual(1)
})

test('edge case: submit saat masih ada soal kosong menampilkan peringatan', async ({ page }) => {
  await installApiMocks(page)

  await loginAsStudent(page)
  await goToWorkspaceAndStartExam(page)
  await selectMcOption(page, 0)
  await page.locator('[data-qnav-idx="2"]').click()
  await page.getByRole('button', { name: 'Selesai', exact: true }).click()

  await expect(page.getByText('Konfirmasi Selesai')).toBeVisible()
  await expect(page.getByText('Masih ada 2 soal belum dijawab.')).toBeVisible()
})

test('edge case: pindah tab memicu heartbeat anti-cheat', async ({ page }) => {
  const state = await installApiMocks(page)

  await loginAsStudent(page)
  await goToWorkspaceAndStartExam(page)

  await Promise.all([
    page.waitForResponse(
      (resp) =>
        resp.url().includes(`/api/v1/student/sessions/${SESSION_ID}/heartbeat`) &&
        resp.request().method() === 'POST',
    ),
    page.evaluate(() => {
      Object.defineProperty(document, 'visibilityState', {
        configurable: true,
        get: () => 'hidden',
      })
      document.dispatchEvent(new Event('visibilitychange'))
    }),
  ])

  expect(state.heartbeatCalls).toBeGreaterThan(0)
})

test('edge case: refresh halaman saat ujian mempertahankan jawaban tersimpan', async ({ page }) => {
  const state = await installApiMocks(page)

  await loginAsStudent(page)
  await goToWorkspaceAndStartExam(page)
  await Promise.all([
    page.waitForResponse(
      (resp) =>
        resp.url().includes(`/api/v1/student/sessions/${SESSION_ID}/answers`) &&
        resp.request().method() === 'POST',
    ),
    selectMcOption(page, 1),
  ])

  await page.reload()
  const tokenPrompt = page.getByRole('heading', { name: 'Masukkan Token Ujian' })
  if (await tokenPrompt.isVisible().catch(() => false)) {
    await page.getByPlaceholder('Masukkan token dari pengawas').fill('TOK123')
    await Promise.all([
      page.waitForResponse(
        (resp) =>
          resp.url().includes(`/api/v1/student/sessions/${SESSION_ID}/verify-token`) &&
          resp.request().method() === 'POST',
      ),
      page.getByRole('button', { name: 'Mulai Ujian' }).click(),
    ])
  }
  await expect(page.getByText('SOAL NOMOR: 1')).toBeVisible({ timeout: 10000 })
  await expect(page.locator('input[name="mc-q-1"]').nth(1)).toBeChecked()
  expect(Object.keys(state.answersByQuestion)).toContain('q-1')
})

test('edge case: mobile viewport exam page render stabil + screenshot', async ({ page }) => {
  await installApiMocks(page)

  await page.setViewportSize({ width: 390, height: 844 })
  await loginAsStudent(page)
  await page.goto('/#/student/ujian')
  await expect(page.getByRole('heading', { name: 'Ruang Ujian' })).toBeVisible()
  await page.screenshot({ path: 'test-results/mobile-student-ujian.png', fullPage: true })
})
