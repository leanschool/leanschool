import { createContext, useCallback, useContext, useEffect, useState } from 'react'
import { computeChallenge, generateRandomString, parseJwtPayload } from './pkce'
import { config } from '../config'

const KEYCLOAK_URL = config.keycloakUrl
const REALM        = config.keycloakRealm
const CLIENT_ID    = config.keycloakClientId
const REDIRECT_URI = window.location.origin

const BASE_URL = `${KEYCLOAK_URL}/realms/${REALM}/protocol/openid-connect`
const TOKEN_KEY = 'leanschool_auth'

// ── token storage ─────────────────────────────────────────────────────────────

function saveTokens(data) {
  const tokens = {
    accessToken:  data.access_token,
    refreshToken: data.refresh_token,
    idToken:      data.id_token,
    expiresAt:    Date.now() + data.expires_in * 1000,
  }
  localStorage.setItem(TOKEN_KEY, JSON.stringify(tokens))
  return tokens
}

function loadTokens() {
  try {
    const raw = localStorage.getItem(TOKEN_KEY)
    return raw ? JSON.parse(raw) : null
  } catch {
    return null
  }
}

function clearTokens() {
  localStorage.removeItem(TOKEN_KEY)
}

// ── Keycloak token exchange helpers ───────────────────────────────────────────

async function exchangeCode(code, verifier) {
  const res = await fetch(`${BASE_URL}/token`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
    body: new URLSearchParams({
      grant_type:    'authorization_code',
      client_id:     CLIENT_ID,
      code,
      redirect_uri:  REDIRECT_URI,
      code_verifier: verifier,
    }),
  })
  if (!res.ok) throw new Error(`token exchange failed: ${res.status}`)
  return res.json()
}

async function refreshAccessToken(refreshToken) {
  const res = await fetch(`${BASE_URL}/token`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
    body: new URLSearchParams({
      grant_type:    'refresh_token',
      client_id:     CLIENT_ID,
      refresh_token: refreshToken,
    }),
  })
  if (!res.ok) throw new Error(`token refresh failed: ${res.status}`)
  return res.json()
}

// ── Auth context ──────────────────────────────────────────────────────────────

const AuthContext = createContext(null)

export function AuthProvider({ children }) {
  const [tokens, setTokens] = useState(() => loadTokens())
  // loading=true while we handle the auth callback redirect
  const [loading, setLoading] = useState(true)

  // On first render: check if Keycloak redirected back with ?code=
  useEffect(() => {
    const params = new URLSearchParams(window.location.search)
    const code  = params.get('code')
    const state = params.get('state')

    if (!code) {
      setLoading(false)
      return
    }

    // Consume the callback URL immediately so a reload doesn't re-trigger.
    window.history.replaceState({}, '', window.location.pathname)

    const verifier   = sessionStorage.getItem('pkce_verifier')
    const savedState = sessionStorage.getItem('pkce_state')
    sessionStorage.removeItem('pkce_verifier')
    sessionStorage.removeItem('pkce_state')

    if (state !== savedState) {
      console.error('[auth] state mismatch — possible CSRF, discarding tokens')
      clearTokens()
      setLoading(false)
      return
    }

    exchangeCode(code, verifier)
      .then(data => setTokens(saveTokens(data)))
      .catch(err => {
        console.error('[auth] code exchange error:', err)
        clearTokens()
      })
      .finally(() => setLoading(false))
  }, [])

  // ── actions ──────────────────────────────────────────────────────────────────

  const login = useCallback(async () => {
    const verifier   = generateRandomString(64)
    const state      = generateRandomString(32)
    const challenge  = await computeChallenge(verifier)

    sessionStorage.setItem('pkce_verifier', verifier)
    sessionStorage.setItem('pkce_state', state)

    const url = new URL(`${BASE_URL}/auth`)
    url.searchParams.set('client_id',             CLIENT_ID)
    url.searchParams.set('response_type',         'code')
    url.searchParams.set('scope',                 'openid profile email')
    url.searchParams.set('redirect_uri',          REDIRECT_URI)
    url.searchParams.set('code_challenge',        challenge)
    url.searchParams.set('code_challenge_method', 'S256')
    url.searchParams.set('state',                 state)

    window.location.href = url.toString()
  }, [])

  const logout = useCallback(() => {
    const idToken = tokens?.idToken
    clearTokens()
    setTokens(null)

    const url = new URL(`${BASE_URL}/logout`)
    url.searchParams.set('client_id',               CLIENT_ID)
    url.searchParams.set('post_logout_redirect_uri', REDIRECT_URI)
    if (idToken) url.searchParams.set('id_token_hint', idToken)

    window.location.href = url.toString()
  }, [tokens])

  // ── authFetch ─────────────────────────────────────────────────────────────────
  // Drop-in replacement for `fetch` that attaches the Bearer token and silently
  // refreshes it when it is within 30 s of expiry.

  const authFetch = useCallback(async (url, options = {}) => {
    let tok = tokens

    if (!tok) {
      login()
      return
    }

    // Proactively refresh if expiring within 30 seconds.
    if (tok.expiresAt - Date.now() < 30_000) {
      try {
        const data = await refreshAccessToken(tok.refreshToken)
        tok = saveTokens(data)
        setTokens(tok)
      } catch {
        clearTokens()
        setTokens(null)
        login()
        return
      }
    }

    return fetch(url, {
      ...options,
      headers: {
        ...options.headers,
        Authorization: `Bearer ${tok.accessToken}`,
      },
    })
  }, [tokens, login])

  const user = tokens ? parseJwtPayload(tokens.accessToken) : null

  return (
    <AuthContext.Provider value={{ tokens, user, loading, login, logout, authFetch }}>
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth() {
  return useContext(AuthContext)
}
