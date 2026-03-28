/**
 * PKCE (Proof Key for Code Exchange) utilities for OAuth2 / OIDC authorization flows.
 */

const CHARS = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-._~'

/** Generate a cryptographically random string suitable for use as a code_verifier or state. */
export function generateRandomString(length = 64) {
  const buf = new Uint8Array(length)
  crypto.getRandomValues(buf)
  return Array.from(buf).map(b => CHARS[b % CHARS.length]).join('')
}

/** Compute the PKCE code_challenge = base64url(sha256(verifier)). */
export async function computeChallenge(verifier) {
  const data = new TextEncoder().encode(verifier)
  const digest = await crypto.subtle.digest('SHA-256', data)
  return btoa(String.fromCharCode(...new Uint8Array(digest)))
    .replace(/\+/g, '-')
    .replace(/\//g, '_')
    .replace(/=/g, '')
}

/** Decode a JWT payload (base64url → JSON). Does NOT verify the signature. */
export function parseJwtPayload(token) {
  try {
    const b64 = token.split('.')[1].replace(/-/g, '+').replace(/_/g, '/')
    const bytes = Uint8Array.from(atob(b64), c => c.charCodeAt(0))
    return JSON.parse(new TextDecoder().decode(bytes))
  } catch {
    return {}
  }
}
