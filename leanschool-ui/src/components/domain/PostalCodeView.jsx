import { useState, useEffect } from 'react'
import { useAuth } from '../../auth/useAuth'
import { useTranslation } from '../../i18n/useTranslation'
import './domain-components.css'

const API = import.meta.env.VITE_LEANSCHOOL_URL || 'http://localhost:8080'

export default function PostalCodeView({ id }) {
  const { t } = useTranslation()
  const { authFetch } = useAuth()
  const [postalCode, setPostalCode] = useState(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)

  useEffect(() => {
    if (id == null) return
    setLoading(true)
    setError(null)
    authFetch(`${API}/postal-codes/${id}`)
      .then(res => {
        if (!res.ok) throw new Error(t.domain.common.loadError)
        return res.json()
      })
      .then(data => setPostalCode(data))
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

  if (!postalCode) return null

  return (
    <div className="dv-card">
      <div className="dv-row">
        <span className="dv-label">{t.domain.postalCode.number}</span>
        <span className="dv-value">{postalCode.number}</span>
      </div>
      <div className="dv-row">
        <span className="dv-label">{t.domain.postalCode.cityName}</span>
        <span className="dv-value">{postalCode.city}</span>
      </div>
      <div className="dv-row">
        <span className="dv-label">{t.domain.common.version}</span>
        <span className="dv-value">{postalCode.version}</span>
      </div>
    </div>
  )
}
