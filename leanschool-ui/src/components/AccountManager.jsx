import { useState, useEffect } from 'react'
import { useTranslation } from '../i18n/useTranslation'
import { useAuth } from '../auth/useAuth'
import './AccountManager.css'

import { config } from '../config'

const LEANSCHOOL_URL = config.leanschoolUrl

const EMPTY_FORM = { name: '', shortcut: '', budget: '', classId: '', validFrom: '', validTo: '' }

export default function AccountManager({ onBack }) {
  const { t } = useTranslation()
  const { authFetch } = useAuth()
  const [accounts, setAccounts] = useState([])
  const [classes, setClasses] = useState([])
  const [loadState, setLoadState] = useState('loading')
  const [creating, setCreating] = useState(false)
  const [createForm, setCreateForm] = useState(EMPTY_FORM)
  const [createError, setCreateError] = useState('')
  const [editId, setEditId] = useState(null)
  const [editForm, setEditForm] = useState(EMPTY_FORM)
  const [editError, setEditError] = useState('')
  const [deleteError, setDeleteError] = useState('')

  const load = () => {
    setLoadState('loading')
    Promise.all([
      authFetch(`${LEANSCHOOL_URL}/accounts`).then(r => r.ok ? r.json() : []),
      authFetch(`${LEANSCHOOL_URL}/school-classes`).then(r => r.ok ? r.json() : []),
    ])
      .then(([accs, cls]) => { setAccounts(accs ?? []); setClasses(cls ?? []); setLoadState('ready') })
      .catch(() => setLoadState('error'))
  }

  useEffect(load, []) // eslint-disable-line react-hooks/exhaustive-deps

  const classMap = Object.fromEntries((classes ?? []).map(c => [c.id, c.name]))

  // ── create ────────────────────────────────────────────────────────────────
  const submitCreate = async e => {
    e.preventDefault()
    setCreateError('')
    try {
      const res = await authFetch(`${LEANSCHOOL_URL}/accounts`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          name: createForm.name,
          shortcut: createForm.shortcut,
          budget: parseFloat(createForm.budget) || 0,
          classId: createForm.classId,
          validFrom: createForm.validFrom ? createForm.validFrom + 'T00:00:00Z' : undefined,
          validTo: createForm.validTo ? createForm.validTo + 'T00:00:00Z' : undefined,
        }),
      })
      if (!res.ok) throw new Error(res.status)
      setCreateForm(EMPTY_FORM)
      setCreating(false)
      load()
    } catch {
      setCreateError(t.accounts.createError)
    }
  }

  // ── edit ──────────────────────────────────────────────────────────────────
  const startEdit = a => {
    setEditId(a.id)
    setEditForm({
      name: a.name,
      shortcut: a.shortcut,
      budget: String(a.budget),
      classId: a.classId ?? '',
      validFrom: a.validFrom ? a.validFrom.slice(0, 10) : '',
      validTo: a.validTo ? a.validTo.slice(0, 10) : '',
    })
    setEditError('')
  }

  const submitEdit = async e => {
    e.preventDefault()
    setEditError('')
    try {
      const res = await authFetch(`${LEANSCHOOL_URL}/accounts/${editId}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          name: editForm.name,
          shortcut: editForm.shortcut,
          budget: parseFloat(editForm.budget) || 0,
          classId: editForm.classId,
          validFrom: editForm.validFrom ? editForm.validFrom + 'T00:00:00Z' : undefined,
          validTo: editForm.validTo ? editForm.validTo + 'T00:00:00Z' : undefined,
        }),
      })
      if (!res.ok) throw new Error(res.status)
      setEditId(null)
      load()
    } catch {
      setEditError(t.accounts.updateError)
    }
  }

  // ── delete ────────────────────────────────────────────────────────────────
  const deleteAccount = async id => {
    setDeleteError('')
    try {
      const res = await authFetch(`${LEANSCHOOL_URL}/accounts/${id}`, { method: 'DELETE' })
      if (!res.ok) throw new Error(res.status)
      load()
    } catch {
      setDeleteError(t.accounts.deleteError)
    }
  }

  return (
    <div className="am-page">
      <div className="orb orb-1" />
      <div className="orb orb-2" />

      <div className="am-header">
        <h2 className="am-title">{t.accounts.title}</h2>
      </div>

      <button className="page-back-btn" onClick={onBack}>← {t.accounts.back}</button>

      <div className="am-content">

        {creating && (
          <form className="am-card am-form" onSubmit={submitCreate}>
            <AccountFormFields form={createForm} onChange={setCreateForm} classes={classes} t={t} autoFocus />
            {createError && <div className="am-error">{createError}</div>}
            <div className="am-form-actions">
              <button type="submit" className="cta-button am-btn">{t.accounts.create}</button>
              <button type="button" className="ghost-button am-btn"
                onClick={() => { setCreating(false); setCreateForm(EMPTY_FORM) }}>
                {t.accounts.cancel}
              </button>
            </div>
          </form>
        )}

        {loadState === 'loading' && (
          <div className="am-center"><div className="spinner" /><span className="am-hint">{t.accounts.loading}</span></div>
        )}
        {loadState === 'error' && (
          <div className="am-center am-error">{t.accounts.error}</div>
        )}
        {loadState === 'ready' && accounts.length === 0 && !creating && (
          <div className="am-center am-hint">{t.accounts.empty}</div>
        )}

        {deleteError && <div className="am-error am-delete-error">{deleteError}</div>}

        {loadState === 'ready' && accounts.map(a => (
          <div key={a.id} className="am-card">
            {editId === a.id ? (
              <form className="am-form" onSubmit={submitEdit}>
                <AccountFormFields form={editForm} onChange={setEditForm} classes={classes} t={t} autoFocus />
                {editError && <div className="am-error">{editError}</div>}
                <div className="am-form-actions">
                  <button type="submit" className="cta-button am-btn">{t.accounts.save}</button>
                  <button type="button" className="ghost-button am-btn" onClick={() => setEditId(null)}>
                    {t.accounts.cancel}
                  </button>
                </div>
              </form>
            ) : (
              <div className="am-account-row">
                <div className="am-account-info">
                  <span className="am-shortcut-badge">{a.shortcut}</span>
                  <div className="am-account-details">
                    <span className="am-account-name">{a.name}</span>
                    {a.classId && classMap[a.classId] && (
                      <span className="am-account-class">{classMap[a.classId]}</span>
                    )}
                  </div>
                </div>
                <div className="am-account-right">
                  <div className="am-budget-info">
                    <span className="am-account-balance">{Number(a.balance).toFixed(2)}</span>
                    <span className="am-budget-detail">
                      {t.accounts.budget}: {Number(a.budget).toFixed(2)} / {t.accounts.spent}: {Number(a.spent).toFixed(2)}
                    </span>
                  </div>
                  <button className="am-icon-btn" onClick={() => startEdit(a)} title={t.accounts.edit}>✎</button>
                  <button className="am-icon-btn am-icon-btn--danger" onClick={() => deleteAccount(a.id)} title={t.accounts.delete}>✕</button>
                </div>
              </div>
            )}
          </div>
        ))}

        {!creating && (
          <button className="am-new-btn am-new-btn--bottom" onClick={() => { setCreating(true); setCreateError('') }}>
            + {t.accounts.newAccount}
          </button>
        )}
      </div>
    </div>
  )
}

function AccountFormFields({ form, onChange, classes, t, autoFocus }) {
  const set = (key, val) => onChange(f => ({ ...f, [key]: val }))
  return (
    <div className="am-form-grid">
      <input
        className="field-input am-field-name"
        placeholder={t.accounts.name}
        value={form.name}
        onChange={e => set('name', e.target.value)}
        required
        autoFocus={autoFocus}
      />
      <input
        className="field-input am-field-shortcut"
        placeholder={t.accounts.shortcut}
        value={form.shortcut}
        onChange={e => set('shortcut', e.target.value.toUpperCase())}
        required
        maxLength={8}
      />
      <input
        className="field-input am-field-budget"
        type="number"
        step="0.01"
        placeholder={t.accounts.budget}
        value={form.budget}
        onChange={e => set('budget', e.target.value)}
      />
      <select
        className="field-input field-select am-field-class"
        value={form.classId}
        onChange={e => set('classId', e.target.value)}
      >
        <option value="">{t.accounts.selectClass}</option>
        {(classes ?? []).map(c => (
          <option key={c.id} value={c.id}>{c.name}</option>
        ))}
      </select>
      <div className="am-date-row">
        <input
          className="field-input am-field-date"
          type="date"
          title={t.accounts.validFrom}
          value={form.validFrom}
          onChange={e => set('validFrom', e.target.value)}
        />
        <input
          className="field-input am-field-date"
          type="date"
          title={t.accounts.validTo}
          value={form.validTo}
          onChange={e => set('validTo', e.target.value)}
        />
      </div>
    </div>
  )
}
