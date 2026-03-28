import { useState, useEffect } from 'react'
import { useAuth } from '../../auth/useAuth'
import { useTranslation } from '../../i18n/useTranslation'
import './domain-components.css'

const API = import.meta.env.VITE_LEANSCHOOL_URL || 'http://localhost:8080'

export default function PostalCodeForm({ id, persist = false, onSave, onCancel }) {
  const { t } = useTranslation()
  const { authFetch } = useAuth()
  const isEdit = id != null

  const [fields, setFields] = useState({ number: '', city: '' })
  const [version, setVersion] = useState(0)
  const [loading, setLoading] = useState(isEdit)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState(null)
  const [conflictError, setConflictError] = useState(null)

  useEffect(() => {
    if (!isEdit) return
    setLoading(true)
    authFetch(`${API}/postal-codes/${id}`)
      .then(res => {
        if (!res.ok) throw new Error(t.domain.common.loadError)
        return res.json()
      })
      .then(data => {
        setFields({ number: String(data.number), city: data.city ?? '' })
        setVersion(data.version)
      })
      .catch(err => setError(err.message))
      .finally(() => setLoading(false))
  }, [id])

  function handleChange(e) {
    const { name, value } = e.target
    setFields(prev => ({ ...prev, [name]: value }))
  }

  async function handleSubmit(e) {
    e.preventDefault()
    setConflictError(null)
    setError(null)

    const body = isEdit
      ? { number: parseInt(fields.number, 10), city: fields.city, version }
      : { number: parseInt(fields.number, 10), city: fields.city }

    if (!persist) {
      onSave?.(body)
      return
    }

    setSaving(true)
    const url = isEdit ? `${API}/postal-codes/${id}` : `${API}/postal-codes`
    const method = isEdit ? 'PUT' : 'POST'

    try {
      const res = await authFetch(url, {
        method,
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body),
      })

      if (res.status === 409) {
        setConflictError(t.domain.common.conflict)
        return
      }
      if (!res.ok) throw new Error(t.domain.common.saveError)

      const saved = await res.json()
      onSave?.(saved)
    } catch (err) {
      setError(err.message)
    } finally {
      setSaving(false)
    }
  }

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

  return (
    <form className="df-form" onSubmit={handleSubmit}>
      <div className="df-grid">
        <div className="df-field-group">
          <div className="df-label">{t.domain.postalCode.number}</div>
          <input
            className="field-input"
            type="number"
            name="number"
            value={fields.number}
            onChange={handleChange}
            placeholder="Postal code number"
            required
          />
        </div>
        <div className="df-field-group">
          <div className="df-label">{t.domain.postalCode.cityName}</div>
          <input
            className="field-input"
            type="text"
            name="city"
            value={fields.city}
            onChange={handleChange}
            placeholder="City name"
            required
          />
        </div>
      </div>

      {conflictError && <div className="df-conflict-error">{conflictError}</div>}

      <div className="df-actions">
        <button type="submit" className="cta-button am-btn" disabled={saving}>
          {saving ? t.domain.common.saving : isEdit ? t.domain.common.update : t.domain.common.create}
        </button>
        {onCancel && (
          <button type="button" className="ghost-button am-btn" onClick={onCancel}>
            {t.domain.common.cancel}
          </button>
        )}
      </div>
    </form>
  )
}
