import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { http, HttpResponse } from 'msw'
import { server } from './mocks/server'
import App from '../App'

// ── Auth mock ─────────────────────────────────────────────────────────────────

const mockUseAuth = vi.hoisted(() => vi.fn())

vi.mock('../auth/useAuth', () => ({
  AuthProvider: ({ children }) => children,
  useAuth: mockUseAuth,
}))

// ── Component stubs (not under test here) ─────────────────────────────────────

vi.mock('../components/Dashboard', () => ({
  default: () => <div data-testid="dashboard">Dashboard</div>,
}))
vi.mock('../components/timetable-planner/TimetablePlanner', () => ({
  default: () => null,
}))
vi.mock('../components/ScanReceipt', () => ({ default: () => null }))
vi.mock('../components/DataDashboard', () => ({ default: () => null }))
vi.mock('../components/TemplateManager', () => ({ default: () => null }))
vi.mock('../components/WhoIsWho', () => ({ default: () => null }))

// ── Helpers ───────────────────────────────────────────────────────────────────

function makeAuthFetch() {
  return vi.fn((url, opts = {}) =>
    fetch(url, { ...opts, headers: { ...opts?.headers, Authorization: 'Bearer test' } })
  )
}

function makeUser(roles = []) {
  return {
    sub: 'u1',
    given_name: 'Ada',
    family_name: 'Lovelace',
    email: 'ada@test.com',
    realm_access: { roles },
  }
}

beforeEach(() => {
  mockUseAuth.mockReturnValue({
    tokens: { accessToken: 'test.token.here', expiresAt: Date.now() + 60_000 },
    user: makeUser(),
    loading: false,
    login: vi.fn(),
    logout: vi.fn(),
    authFetch: makeAuthFetch(),
  })
})

// ── Tests ─────────────────────────────────────────────────────────────────────

describe('registration gate', () => {
  it('shows RegistrationWizard for a new user with no registration', async () => {
    // default MSW handler returns { registrationStatus: 'none' }
    render(<App />)
    expect(await screen.findByText('Request Access')).toBeInTheDocument()
  })

  it('shows AwaitingApproval for a pending user', async () => {
    server.use(
      http.get('http://localhost:8080/users/me', () =>
        HttpResponse.json({ registrationStatus: 'pending' })
      )
    )
    render(<App />)
    expect(await screen.findByText('Awaiting Approval')).toBeInTheDocument()
  })

  it('shows DeniedScreen for a denied user', async () => {
    server.use(
      http.get('http://localhost:8080/users/me', () =>
        HttpResponse.json({ registrationStatus: 'denied' })
      )
    )
    render(<App />)
    expect(await screen.findByText('Access Denied')).toBeInTheDocument()
  })

  it('shows the dashboard for a user with a business role', async () => {
    mockUseAuth.mockReturnValue({
      tokens: { accessToken: 'test.token.here', expiresAt: Date.now() + 60_000 },
      user: makeUser(['teacher']),
      loading: false,
      login: vi.fn(),
      logout: vi.fn(),
      authFetch: makeAuthFetch(),
    })
    server.use(
      http.get('http://localhost:8080/users/me', () =>
        HttpResponse.json({ registrationStatus: 'approved', profileComplete: true })
      )
    )
    render(<App />)
    expect(await screen.findByTestId('dashboard')).toBeInTheDocument()
    expect(screen.queryByText('Request Access')).not.toBeInTheDocument()
  })

  it('shows the loading screen while /users/me is in flight', async () => {
    // statusLoading initialises to true (tokens exist), so loading shows synchronously
    render(<App />)
    expect(screen.getByText('Loading…')).toBeInTheDocument()
    // wait for it to resolve so we don't leave pending requests
    await screen.findByText('Request Access')
  })

  it('shows an error screen when /users/me fails', async () => {
    server.use(
      http.get('http://localhost:8080/users/me', () =>
        new HttpResponse(null, { status: 500 })
      )
    )
    render(<App />)
    expect(await screen.findByText('Unable to reach server.')).toBeInTheDocument()
    expect(screen.queryByText('Request Access')).not.toBeInTheDocument()
  })

  it('re-fetches and shows the wizard after clicking Retry', async () => {
    const user = userEvent.setup()

    // First attempt fails
    server.use(
      http.get('http://localhost:8080/users/me', () =>
        new HttpResponse(null, { status: 500 })
      )
    )
    render(<App />)
    await screen.findByText('Unable to reach server.')

    // Second attempt succeeds
    server.use(
      http.get('http://localhost:8080/users/me', () =>
        HttpResponse.json({ registrationStatus: 'none' })
      )
    )
    await user.click(screen.getByRole('button', { name: /retry/i }))
    expect(await screen.findByText('Request Access')).toBeInTheDocument()
  })
})
