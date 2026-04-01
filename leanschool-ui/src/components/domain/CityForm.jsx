import { useState, useEffect } from 'react'
import { useAuth } from '../../auth/useAuth'
import { useTranslation } from '../../i18n/useTranslation'
import './domain-components.css'

import { config } from '../../config'

const API = config.leanschoolUrl

export default function CityForm({ id, persist = false, onSave, onCancel }) {
  const { t } = useTranslation()
  const { authFetch } = useAuth()
  const isEdit = id != null

  const [name, setName] = useState('')
  const [version, setVersion] = useState(0)
  const [selectedPostalCodes, setSelectedPostalCodes] = useState([])
  const [allPostalCodes, setAllPostalCodes] = useState([])

  const [loading, setLoading] = useState(isEdit)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState(null)
  const [conflictError, setConflictError] = useState(null)

  // Load all postal codes for the multi-select
  useEffect(() => {
    authFetch(`${API}/postal-codes`)
      .then(res => res.ok ? res.json() : Promise.reject(new Error(t.domain.common.loadError)))
      .then(data => setAllPostalCodes(Array.isArray(data) ? data : []))
      .catch(err => setError(err.message))
  }, [])

  // Load existing city when editing
  useEffect(() => {
    if (!isEdit) return
    setLoading(true)
    authFetch(`${API}/cities/${id}`)
      .then(res => {
        if (!res.ok) throw new Error(t.domain.common.loadError)
        return res.json()
      })
      .then(data => {
        setName(data.name ?? '')
        setVersion(data.version)
        setSelectedPostalCodes((data.postalCodes ?? []).map(pc => pc.id))
      })
      .catch(err => setError(err.message))
      .finally(() => setLoading(false))
  }, [id])

  function togglePostalCode(id) {
    setSelectedPostalCodes(prev =>
      prev.includes(id) ? prev.filter(n => n !== id) : [...prev, id]
    )
  }

  async function handleSubmit(e) {
    e.preventDefault()
    setConflictError(null)
    setError(null)

    const postalCodes = allPostalCodes.filter(pc => selectedPostalCodes.includes(pc.id))
    const body = isEdit
      ? { id, name, postalCodes, version }
      : { name, postalCodes }

    if (!persist) {
      onSave?.(body)
      return
    }

    setSaving(true)
    const url = isEdit ? `${API}/cities/${id}` : `${API}/cities`
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
          <div className="df-label">{t.domain.city.name}</div>
          <input
            className="field-input"
            type="text"
            value={name}
            onChange={e => setName(e.target.value)}
            placeholder="City name"
            required
          />
        </div>
      </div>

      <div className="df-section">
        <div className="df-section-title">{t.domain.city.postalCodes}</div>
        {allPostalCodes.length === 0 ? (
          <span style={{ fontSize: 13, color: 'var(--text-3)' }}>No postal codes available</span>
        ) : (
          <div style={{ display: 'flex', flexWrap: 'wrap', gap: 8 }}>
            {allPostalCodes.map(pc => {
              const checked = selectedPostalCodes.includes(pc.id)
              return (
                <label
                  key={pc.id}
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: 6,
                    fontSize: 13,
                    cursor: 'pointer',
                    padding: '4px 10px',
                    borderRadius: 8,
                    border: `1px solid ${checked ? 'var(--accent)' : 'var(--border-1)'}`,
                    background: checked ? 'var(--accent-a8)' : 'transparent',
                    color: checked ? 'var(--accent)' : 'var(--text-2)',
                    transition: 'all 0.15s',
                  }}
                >
                  <input
                    type="checkbox"
                    checked={checked}
                    onChange={() => togglePostalCode(pc.id)}
                    style={{ display: 'none' }}
                  />
                  {pc.number} {pc.city}
                </label>
              )
            })}
          </div>
        )}
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
