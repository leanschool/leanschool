import { useState, useEffect } from 'react'
import { useAuth } from '../../auth/useAuth'
import { useTranslation } from '../../i18n/useTranslation'
import './domain-components.css'

import { config } from '../../config'

const API = config.leanschoolUrl

const EMPTY = { name: '', shortcut: '', budget: '', classId: '', validFrom: '', validTo: '' }

export default function AccountForm({ id, persist = false, onSave, onCancel }) {
  const { t } = useTranslation()
  const { authFetch } = useAuth()
  const [fields, setFields] = useState(EMPTY)
  const [classes, setClasses] = useState([])
  const [loading, setLoading] = useState(!!id)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState(null)
  const [conflict, setConflict] = useState(false)

  useEffect(() => {
    authFetch(`${API}/school-classes`)
      .then(res => res.ok ? res.json() : [])
      .then(data => setClasses(data ?? []))
      .catch(() => {})
  }, []) // eslint-disable-line react-hooks/exhaustive-deps

  useEffect(() => {
    if (!id) return
    setLoading(true)
    authFetch(`${API}/accounts/${id}`)
      .then(res => {
        if (!res.ok) throw new Error(`Failed to load account (${res.status})`)
        return res.json()
      })
      .then(data => {
        setFields({
          name: data.name ?? '',
          shortcut: data.shortcut ?? '',
          budget: data.budget != null ? String(data.budget) : '',
          classId: data.classId ?? '',
          validFrom: data.validFrom ? data.validFrom.slice(0, 10) : '',
          validTo: data.validTo ? data.validTo.slice(0, 10) : '',
        })
      })
      .catch(err => setError(err.message))
      .finally(() => setLoading(false))
  }, [id]) // eslint-disable-line react-hooks/exhaustive-deps

  function handleChange(e) {
    const { name, value } = e.target
    setFields(prev => ({ ...prev, [name]: name === 'shortcut' ? value.toUpperCase() : value }))
  }

  async function handleSubmit(e) {
    e.preventDefault()
    setConflict(false)
    setError(null)

    const body = {
      name: fields.name,
      shortcut: fields.shortcut,
      budget: parseFloat(fields.budget) || 0,
      ...(fields.classId ? { classId: fields.classId } : {}),
      ...(fields.validFrom ? { validFrom: fields.validFrom + 'T00:00:00Z' } : {}),
      ...(fields.validTo ? { validTo: fields.validTo + 'T00:00:00Z' } : {}),
    }

    if (!persist) {
      onSave?.(body)
      return
    }

    setSaving(true)
    const url = id ? `${API}/accounts/${id}` : `${API}/accounts`
    const method = id ? 'PUT' : 'POST'

    try {
      const res = await authFetch(url, {
        method,
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body),
      })

      if (res.status === 409) {
        setConflict(true)
        return
      }
      if (!res.ok) throw new Error(`Save failed (${res.status})`)

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

  return (
    <form className="df-form" onSubmit={handleSubmit}>
      <div className="df-grid">
        <div className="df-field-group">
          <div className="df-label">{t.domain.account.name}</div>
          <input
            className="field-input"
            type="text"
            name="name"
            value={fields.name}
            onChange={handleChange}
            required
          />
        </div>
        <div className="df-field-group">
          <div className="df-label">{t.domain.account.shortcut}</div>
          <input
            className="field-input"
            type="text"
            name="shortcut"
            value={fields.shortcut}
            onChange={handleChange}
            required
            maxLength={8}
          />
        </div>
        <div className="df-field-group">
          <div className="df-label">{t.domain.account.budget}</div>
          <input
            className="field-input"
            type="number"
            step="0.01"
            name="budget"
            value={fields.budget}
            onChange={handleChange}
          />
        </div>
        <div className="df-field-group">
          <div className="df-label">{t.domain.account.class}</div>
          <select
            className="field-input field-select"
            name="classId"
            value={fields.classId}
            onChange={handleChange}
          >
            <option value="">{t.domain.common.none}</option>
            {classes.map(c => (
              <option key={c.id} value={c.id}>{c.name}</option>
            ))}
          </select>
        </div>
        <div className="df-field-group">
          <div className="df-label">{t.domain.account.validFrom}</div>
          <input
            className="field-input"
            type="date"
            name="validFrom"
            value={fields.validFrom}
            onChange={handleChange}
          />
        </div>
        <div className="df-field-group">
          <div className="df-label">{t.domain.account.validTo}</div>
          <input
            className="field-input"
            type="date"
            name="validTo"
            value={fields.validTo}
            onChange={handleChange}
          />
        </div>
      </div>

      {conflict && (
        <div className="df-conflict-error">
          {t.domain.common.conflict}
        </div>
      )}
      {error && <div className="dv-error">{error}</div>}

      <div className="df-actions">
        <button type="submit" className="cta-button am-btn" disabled={saving}>
          {saving ? t.domain.common.saving : id ? t.domain.common.update : t.domain.common.create}
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
