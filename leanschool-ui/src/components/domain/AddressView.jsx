import { useState, useEffect } from 'react'
import { useAuth } from '../../auth/useAuth'
import { useTranslation } from '../../i18n/useTranslation'
import './domain-components.css'

const API = import.meta.env.VITE_LEANSCHOOL_URL || 'http://localhost:8080'

export default function AddressView({ id }) {
  const { t } = useTranslation()
  const { authFetch } = useAuth()
  const [address, setAddress] = useState(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)

  useEffect(() => {
    if (id == null) return
    setLoading(true)
    setError(null)
    authFetch(`${API}/addresses/${id}`)
      .then(res => {
        if (!res.ok) throw new Error(t.domain.common.loadError)
        return res.json()
      })
      .then(data => setAddress(data))
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

  if (!address) return null

  const pc = address.postalCode

  return (
    <div className="dv-card">
      <div className="dv-row">
        <span className="dv-label">{t.domain.common.id}</span>
        <span className="dv-value">{address.id}</span>
      </div>
      {address.street != null && (
        <div className="dv-row">
          <span className="dv-label">{t.domain.address.street}</span>
          <span className="dv-value">{address.street}</span>
        </div>
      )}
      {address.number != null && (
        <div className="dv-row">
          <span className="dv-label">{t.domain.address.houseNumber}</span>
          <span className="dv-value">{address.number}</span>
        </div>
      )}
      <div className="dv-row">
        <span className="dv-label">{t.domain.address.postalCode}</span>
        <span className="dv-value">
          {pc ? `${pc.number} ${pc.city}` : '—'}
        </span>
      </div>
      <div className="dv-row">
        <span className="dv-label">{t.domain.common.version}</span>
        <span className="dv-value">{address.version}</span>
      </div>
    </div>
  )
}
