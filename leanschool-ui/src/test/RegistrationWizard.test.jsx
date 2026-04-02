import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { http, HttpResponse } from 'msw'
import { server } from './mocks/server'
import { I18nProvider } from '../i18n/useTranslation'
import RegistrationWizard from '../components/RegistrationWizard'

// ── Auth mock ─────────────────────────────────────────────────────────────────

const mockUseAuth = vi.hoisted(() => vi.fn())

vi.mock('../auth/useAuth', () => ({
  AuthProvider: ({ children }) => children,
  useAuth: mockUseAuth,
}))

// ── Helpers ───────────────────────────────────────────────────────────────────

function renderWizard(onRegistered = vi.fn()) {
  return render(
    <I18nProvider>
      <RegistrationWizard onRegistered={onRegistered} />
    </I18nProvider>
  )
}

beforeEach(() => {
  const authFetch = vi.fn((url, opts = {}) =>
    fetch(url, { ...opts, headers: { ...opts?.headers, Authorization: 'Bearer test' } })
  )
  mockUseAuth.mockReturnValue({
    tokens: { accessToken: 'test.token.here' },
    user: {
      sub: 'u1',
      given_name: 'Ada',
      family_name: 'Lovelace',
      email: 'ada@test.com',
      realm_access: { roles: [] },
    },
    loading: false,
    login: vi.fn(),
    logout: vi.fn(),
    authFetch,
  })
})

// ── Step 1: Role Selection ────────────────────────────────────────────────────

describe('Step 1 — role selection', () => {
  it('loads roles from the API and renders checkboxes', async () => {
    renderWizard()
    expect(await screen.findByRole('checkbox', { name: /teacher/i })).toBeInTheDocument()
    expect(screen.getByRole('checkbox', { name: /school management/i })).toBeInTheDocument()
  })

  it('does not advance when Next is clicked with no role selected', async () => {
    const user = userEvent.setup()
    renderWizard()
    await screen.findByRole('checkbox', { name: /teacher/i })
    await user.click(screen.getByRole('button', { name: /next/i }))
    // still on step 1
    expect(screen.getByText('My roles')).toBeInTheDocument()
    expect(screen.queryByText('Personal Information')).not.toBeInTheDocument()
  })

  it('advances to step 2 after selecting a role', async () => {
    const user = userEvent.setup()
    renderWizard()
    await user.click(await screen.findByRole('checkbox', { name: /school management/i }))
    await user.click(screen.getByRole('button', { name: /next/i }))
    expect(await screen.findByText('Personal Information')).toBeInTheDocument()
  })

  it('allows selecting multiple roles', async () => {
    const user = userEvent.setup()
    renderWizard()
    await user.click(await screen.findByRole('checkbox', { name: /teacher/i }))
    await user.click(screen.getByRole('checkbox', { name: /school management/i }))
    expect(screen.getByRole('checkbox', { name: /teacher/i })).toBeChecked()
    expect(screen.getByRole('checkbox', { name: /school management/i })).toBeChecked()
  })
})

// ── Step 2: Personal Info ─────────────────────────────────────────────────────

describe('Step 2 — personal info', () => {
  async function goToStep2() {
    const user = userEvent.setup()
    renderWizard()
    await user.click(await screen.findByRole('checkbox', { name: /school management/i }))
    await user.click(screen.getByRole('button', { name: /next/i }))
    await screen.findByText('Personal Information')
    return user
  }

  it('pre-fills the email field from the JWT user claim', async () => {
    await goToStep2()
    expect(screen.getByRole('textbox', { name: /email/i })).toHaveValue('ada@test.com')
  })

  it('does not advance when both name fields are empty', async () => {
    const user = await goToStep2()
    await user.click(screen.getByRole('button', { name: /next/i }))
    expect(screen.getByText('Personal Information')).toBeInTheDocument()
  })

  it('advances to step 3 when at least a first name is provided', async () => {
    const user = await goToStep2()
    await user.type(screen.getByRole('textbox', { name: /first name/i }), 'Ada')
    await user.click(screen.getByRole('button', { name: /next/i }))
    expect(await screen.findByText('Role-Specific Details')).toBeInTheDocument()
  })
})

// ── Step 3: Role Details ──────────────────────────────────────────────────────

describe('Step 3 — role details', () => {
  async function goToStep3AsTeacher() {
    const user = userEvent.setup()
    renderWizard()
    await user.click(await screen.findByRole('checkbox', { name: /^teacher$/i }))
    await user.click(screen.getByRole('button', { name: /next/i }))
    await user.type(await screen.findByRole('textbox', { name: /first name/i }), 'Ada')
    await user.click(screen.getByRole('button', { name: /next/i }))
    await screen.findByText('Role-Specific Details')
    return user
  }

  it('shows school class checkboxes when teacher role is selected', async () => {
    await goToStep3AsTeacher()
    expect(await screen.findByRole('checkbox', { name: /1a/i })).toBeInTheDocument()
    expect(screen.getByRole('checkbox', { name: /2b/i })).toBeInTheDocument()
  })

  it('allows toggling a class checkbox', async () => {
    const user = await goToStep3AsTeacher()
    const checkbox = await screen.findByRole('checkbox', { name: /1a/i })
    await user.click(checkbox)
    expect(checkbox).toBeChecked()
    await user.click(checkbox)
    expect(checkbox).not.toBeChecked()
  })

  it('shows no class checkboxes for a non-teacher role', async () => {
    const user = userEvent.setup()
    renderWizard()
    await user.click(await screen.findByRole('checkbox', { name: /school management/i }))
    await user.click(screen.getByRole('button', { name: /next/i }))
    await user.type(await screen.findByRole('textbox', { name: /first name/i }), 'Ada')
    await user.click(screen.getByRole('button', { name: /next/i }))
    await screen.findByText('Role-Specific Details')
    expect(screen.queryByRole('checkbox', { name: /1a/i })).not.toBeInTheDocument()
  })
})

// ── Step 4: Confirmation & Submit ─────────────────────────────────────────────

describe('Step 4 — confirmation and submit', () => {
  async function goToStep4() {
    const user = userEvent.setup()
    renderWizard(vi.fn())
    await user.click(await screen.findByRole('checkbox', { name: /school management/i }))
    await user.click(screen.getByRole('button', { name: /next/i }))
    await user.type(await screen.findByRole('textbox', { name: /first name/i }), 'Ada')
    await user.type(screen.getByRole('textbox', { name: /last name/i }), 'Lovelace')
    await user.click(screen.getByRole('button', { name: /next/i }))
    await screen.findByText('Role-Specific Details')
    await user.click(screen.getByRole('button', { name: /next/i }))
    await screen.findByText('Confirm Your Request')
    return user
  }

  it('shows a summary of the collected data', async () => {
    await goToStep4()
    expect(screen.getByText('Ada')).toBeInTheDocument()
    expect(screen.getByText('Lovelace')).toBeInTheDocument()
    expect(screen.getByText('ada@test.com')).toBeInTheDocument()
  })

  it('sends the correct request body on submit', async () => {
    let capturedBody = null
    server.use(
      http.post('http://localhost:8080/registration/start', async ({ request }) => {
        capturedBody = await request.json()
        return new HttpResponse(null, { status: 200 })
      })
    )
    const user = await goToStep4()
    await user.click(screen.getByRole('button', { name: /submit request/i }))
    await screen.findByText('Request Access') // onRegistered triggers status→pending; re-render not needed here

    expect(capturedBody).toMatchObject({
      desiredRoles: ['school-management'],
      personData: { prename: 'Ada', name: 'Lovelace' },
      contactEmail: 'ada@test.com',
    })
  })

  it('calls onRegistered after a successful submit', async () => {
    const onRegistered = vi.fn()
    const user = userEvent.setup()
    render(
      <I18nProvider>
        <RegistrationWizard onRegistered={onRegistered} />
      </I18nProvider>
    )
    await user.click(await screen.findByRole('checkbox', { name: /school management/i }))
    await user.click(screen.getByRole('button', { name: /next/i }))
    await user.type(await screen.findByRole('textbox', { name: /first name/i }), 'Ada')
    await user.click(screen.getByRole('button', { name: /next/i }))
    await screen.findByText('Role-Specific Details')
    await user.click(screen.getByRole('button', { name: /next/i }))
    await screen.findByText('Confirm Your Request')
    await user.click(screen.getByRole('button', { name: /submit request/i }))
    // wait for the async submit to complete
    await vi.waitFor(() => expect(onRegistered).toHaveBeenCalledOnce())
  })

  it('shows an error message when the submit request fails', async () => {
    server.use(
      http.post('http://localhost:8080/registration/start', () =>
        new HttpResponse('Internal Server Error', { status: 500 })
      )
    )
    const user = await goToStep4()
    await user.click(screen.getByRole('button', { name: /submit request/i }))
    expect(await screen.findByText('Internal Server Error')).toBeInTheDocument()
    // wizard stays open
    expect(screen.getByText('Confirm Your Request')).toBeInTheDocument()
  })
})
