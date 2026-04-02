import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { I18nProvider } from '../i18n/useTranslation'
import DeniedScreen from '../components/DeniedScreen'

// ── Auth mock ─────────────────────────────────────────────────────────────────

const mockUseAuth = vi.hoisted(() => vi.fn())

vi.mock('../auth/useAuth', () => ({
  AuthProvider: ({ children }) => children,
  useAuth: mockUseAuth,
}))

const mockLogout = vi.fn()

beforeEach(() => {
  mockLogout.mockReset()
  mockUseAuth.mockReturnValue({
    tokens: { accessToken: 'test.token.here' },
    user: { sub: 'u1' },
    loading: false,
    login: vi.fn(),
    logout: mockLogout,
    authFetch: vi.fn(),
  })
})

// ── Tests ─────────────────────────────────────────────────────────────────────

describe('DeniedScreen', () => {
  it('renders the denied title and subtitle', () => {
    render(<I18nProvider><DeniedScreen /></I18nProvider>)
    expect(screen.getByText('Access Denied')).toBeInTheDocument()
    expect(screen.getByText(/not approved/i)).toBeInTheDocument()
  })

  it('calls logout when the Back button is clicked', async () => {
    const user = userEvent.setup()
    render(<I18nProvider><DeniedScreen /></I18nProvider>)
    await user.click(screen.getByRole('button', { name: /back/i }))
    expect(mockLogout).toHaveBeenCalledOnce()
  })
})
