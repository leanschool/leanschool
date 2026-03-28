import { useState, useEffect } from 'react'
import { useAuth } from '../../auth/useAuth'
import { useTranslation } from '../../i18n/useTranslation'
import './domain-components.css'

const API = import.meta.env.VITE_LEANSCHOOL_URL || 'http://localhost:8080'

export default function CityView({ id }) {
  const { t } = useTranslation()
  const { authFetch } = useAuth()
  const [city, setCity] = useState(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)

  useEffect(() => {
    if (id == null) return
    setLoading(true)
    setError(null)
    authFetch(`${API}/cities/${id}`)
      .then(res => {
        if (!res.ok) throw new Error(t.domain.common.loadError)
        return res.json()
      })
      .then(data => setCity(data))
      .catch(err => setError(err.message))
      .finally(() => setLoading(false))
  }, [id])

  if (loading) {
    return (
      <div className="dv-loading">
        <div className="spinner" />
      </div>
    )
  }

  if (error) {
    return <div className="dv-error">{error}</div>
  }

  if (!city) return null

  const postalCodes = city.postalCodes ?? []

  return (
    <div className="dv-card">
      <div className="dv-row">
        <span className="dv-label">{t.domain.common.id}</span>
        <span className="dv-value">{city.id}</span>
      </div>
      <div className="dv-row">
        <span className="dv-label">{t.domain.city.name}</span>
        <span className="dv-value">{city.name}</span>
      </div>
      <div className="dv-row">
        <span className="dv-label">{t.domain.common.version}</span>
        <span className="dv-value">{city.version}</span>
      </div>

      <div className="dv-section-title">{t.domain.city.postalCodes}</div>
      {postalCodes.length === 0 ? (
        <span className="dv-value">{t.domain.common.none}</span>
      ) : (
        <div className="dv-badge-list">
          {postalCodes.map(pc => (
            <span key={pc.id} className="dv-badge">
              {pc.number} {pc.city}
            </span>
          ))}
        </div>
      )}
    </div>
  )
}
