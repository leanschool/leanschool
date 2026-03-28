import createClient from 'openapi-fetch'
import { useAuth } from '../auth/useAuth'

const BASE_URL = import.meta.env.VITE_LEANSCHOOL_URL || 'http://localhost:8080'

/**
 * Returns a type-safe openapi-fetch client for the leanschool API,
 * with the Keycloak Bearer token automatically injected on every request.
 *
 * Usage:
 *   const api = useLeanschoolApi()
 *   const { data, error } = await api.GET('/locations/{id}', { params: { path: { id } } })
 */
export function useLeanschoolApi() {
  const { tokens, login } = useAuth()

  const client = createClient({ baseUrl: BASE_URL })

  client.use({
    async onRequest({ request }) {
      if (!tokens) {
        login()
        return request
      }
      request.headers.set('Authorization', `Bearer ${tokens.accessToken}`)
      return request
    },
  })

  return client
}
