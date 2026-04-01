import { useState, useEffect } from 'react'
import { useAuth } from '../../auth/useAuth'
import { useTranslation } from '../../i18n/useTranslation'
import './domain-components.css'

import { config } from '../../config'

const API = config.leanschoolUrl

export default function AddressForm({ id, persist = false, onSave, onCancel }) {
  const { t } = useTranslation()
  const { authFetch } = useAuth()
  const isEdit = id != null

  const [fields, setFields] = useState({ street: '', number: '' })
  const [version, setVersion] = useState(0)
  const [selectedPostalCode, setSelectedPostalCode] = useState('')
  const [allPostalCodes, setAllPostalCodes] = useState([])

  const [loading, setLoading] = useState(isEdit)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState(null)
  const [conflictError, setConflictError] = useState(null)

  // Load postal codes for the select
  useEffect(() => {
    authFetch(`${API}/postal-codes`)
      .then(res => res.ok ? res.json() : Promise.reject(new Error(t.domain.common.loadError)))
      .then(data => setAllPostalCodes(Array.isArray(data) ? data : []))
      .catch(err => setError(err.message))
  }, [])

  // Load existing address when editing
  useEffect(() => {
    if (!isEdit) return
    setLoading(true)
    authFetch(`${API}/addresses/${id}`)
      .then(res => {
        if (!res.ok) throw new Error(t.domain.common.loadError)
        return res.json()
      })
      .then(data => {
        setFields({
          street: data.street ?? '',
          number: data.number ?? '',
        })
        setVersion(data.version)
        setSelectedPostalCode(data.postalCode ? data.postalCode.id : '')
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

    const postalCode = selectedPostalCode
      ? allPostalCodes.find(pc => pc.id === selectedPostalCode) ?? null
      : null

    const body = {
      ...(isEdit ? { id, version } : {}),
      ...(fields.street !== '' ? { street: fields.street } : {}),
      ...(fields.number !== '' ? { number: fields.number } : {}),
      ...(postalCode ? { postalCode } : {}),
    }

    if (!persist) {
      onSave?.(body)
      return
    }

    setSaving(true)
    const url = isEdit ? `${API}/addresses/${id}` : `${API}/addresses`
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
          <div className="df-label">{t.domain.address.street}</div>
          <input
            className="field-input"
            type="text"
            name="street"
            value={fields.street}
            onChange={handleChange}
            placeholder="Street name (optional)"
          />
        </div>
        <div className="df-field-group">
          <div className="df-label">{t.domain.address.houseNumber}</div>
          <input
            className="field-input"
            type="text"
            name="number"
            value={fields.number}
            onChange={handleChange}
            placeholder="House number (optional)"
          />
        </div>
        <div className="df-field-group">
          <div className="df-label">{t.domain.address.postalCode}</div>
          <select
            className="field-input field-select"
            value={selectedPostalCode}
            onChange={e => setSelectedPostalCode(e.target.value)}
          >
            <option value="">{t.domain.common.none}</option>
            {allPostalCodes.map(pc => (
              <option key={pc.id} value={pc.id}>
                {pc.number} — {pc.city}
              </option>
            ))}
          </select>
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
