import { useState, useEffect } from 'react'
import { useTranslation } from '../i18n/useTranslation'
import { useAuth } from '../auth/useAuth'
import AuthImage from './AuthImage'
import './ReceiptForm.css'

const LEANSCHOOL_URL   = import.meta.env.VITE_LEANSCHOOL_URL   || 'http://localhost:8080'
const FILE_SERVICE_URL = import.meta.env.VITE_FILE_SERVICE_URL || 'http://localhost:8083'

export default function ReceiptForm({ receipt: initial, onSaved, onCancel }) {
  const { t } = useTranslation()
  const { authFetch, user } = useAuth()
  const [receipt, setReceipt] = useState(() => normalise(initial, user?.sub))
  const [state, setState] = useState('idle') // idle | sending | error
  const [errorMsg, setErrorMsg] = useState('')
  const [accounts, setAccounts] = useState([])
  const [classes, setClasses] = useState([])
  const [dataLoading, setDataLoading] = useState(true)

  useEffect(() => {
    Promise.all([
      authFetch(`${LEANSCHOOL_URL}/accounts`).then(r => r.ok ? r.json() : []),
      authFetch(`${LEANSCHOOL_URL}/school-classes`).then(r => r.ok ? r.json() : []),
    ])
      .then(([accs, cls]) => { setAccounts(accs ?? []); setClasses(cls ?? []) })
      .catch(() => {})
      .finally(() => setDataLoading(false))
  }, []) // eslint-disable-line react-hooks/exhaustive-deps

  // ── field helpers ────────────────────────────────────────────────────────
  const setField = (key, value) => setReceipt(r => ({ ...r, [key]: value }))

  const setItem = (i, key, value) =>
    setReceipt(r => {
      const items = [...r.items]
      items[i] = { ...items[i], [key]: value }
      return { ...r, items }
    })

  const addItem = () =>
    setReceipt(r => ({ ...r, items: [...r.items, { name: '', amount: 1, price: 0 }] }))

  const removeItem = i =>
    setReceipt(r => ({ ...r, items: r.items.filter((_, idx) => idx !== i) }))

  // ── split helpers ────────────────────────────────────────────────────────
  const addSplit = () =>
    setReceipt(r => ({ ...r, splits: [...r.splits, { classId: '', accountId: '', amount: '' }] }))

  const setSplit = (i, key, value) =>
    setReceipt(r => {
      const splits = [...r.splits]
      // When class changes, clear accountId so user re-selects
      if (key === 'classId') {
        splits[i] = { ...splits[i], classId: value, accountId: '' }
      } else {
        splits[i] = { ...splits[i], [key]: value }
      }
      return { ...r, splits }
    })

  const removeSplit = i =>
    setReceipt(r => ({ ...r, splits: r.splits.filter((_, idx) => idx !== i) }))

  const accountsForClass = classId =>
    classId ? accounts.filter(a => a.classId === classId) : accounts

  const totalPrice = parseFloat(receipt.totalPrice) || 0
  const allocated = receipt.splits.reduce((s, sp) => s + (parseFloat(sp.amount) || 0), 0)
  const remaining = totalPrice - allocated

  const canSubmit = receipt.splits.length > 0
    ? receipt.splits.every(sp => sp.accountId && parseFloat(sp.amount) > 0)
    : false

  // ── submit ───────────────────────────────────────────────────────────────
  const handleSubmit = async e => {
    e.preventDefault()
    setState('sending')
    setErrorMsg('')

    try {
      const payload = {
        ...receipt,
        totalPrice: parseFloat(receipt.totalPrice) || 0,
        taxes: parseFloat(receipt.taxes) || 0,
        items: receipt.items.map(it => ({
          name: it.name,
          amount: parseFloat(it.amount) || 0,
          price: parseFloat(it.price) || 0,
        })),
        splits: receipt.splits.map(sp => ({
          classId: sp.classId,
          accountId: sp.accountId,
          amount: parseFloat(sp.amount) || 0,
        })),
        accountId: receipt.splits[0]?.accountId ?? '',
      }

      const res = await authFetch(`${LEANSCHOOL_URL}/receipts`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
      })

      if (!res.ok) throw new Error(`${res.status}`)

      const saved = await res.json()
      onSaved(saved)
    } catch (err) {
      console.error(err)
      setState('error')
      setErrorMsg(t.form.error)
    }
  }

  return (
    <div className="form-page">
      <div className="orb orb-1" />
      <div className="orb orb-2" />

      <div className="form-layout">
      <form className="form-card" onSubmit={handleSubmit} noValidate>
        <div className="form-header">
          <h2 className="form-title">{t.form.title}</h2>
          <p className="form-subtitle">{t.form.subtitle}</p>
        </div>

        {state === 'error' && (
          <div className="form-error">{errorMsg}</div>
        )}

        {/* owner — read-only */}
        <div className="field-group">
          <label className="field-label">{t.form.owner}</label>
          <div className="field-owner-display">
            <span className="field-owner-name">{user?.preferred_username ?? user?.sub}</span>
            <span className="field-owner-sub">{receipt.receiptOwnerId}</span>
          </div>
        </div>

        {/* date */}
        <div className="field-group">
          <label className="field-label">{t.form.date}</label>
          <input
            className="field-input"
            type="datetime-local"
            value={toDatetimeLocal(receipt.time)}
            onChange={e => setField('time', e.target.value)}
          />
        </div>

        {/* totals row */}
        <div className="field-row">
          <div className="field-group">
            <label className="field-label">{t.form.totalPrice}</label>
            <input
              className="field-input"
              type="number"
              step="0.01"
              min="0"
              value={receipt.totalPrice}
              onChange={e => setField('totalPrice', e.target.value)}
            />
          </div>
          <div className="field-group">
            <label className="field-label">{t.form.taxes}</label>
            <input
              className="field-input"
              type="number"
              step="0.01"
              min="0"
              value={receipt.taxes}
              onChange={e => setField('taxes', e.target.value)}
            />
          </div>
        </div>

        {/* items */}
        <div className="field-group">
          <label className="field-label">{t.form.items}</label>
          <div className="items-list">
            {receipt.items.map((item, i) => (
              <div key={i} className="item-row">
                <input
                  className="field-input item-name"
                  type="text"
                  placeholder={t.form.item.name}
                  value={item.name}
                  onChange={e => setItem(i, 'name', e.target.value)}
                />
                <input
                  className="field-input item-amount"
                  type="number"
                  step="0.001"
                  min="0"
                  placeholder={t.form.item.amount}
                  value={item.amount}
                  onChange={e => setItem(i, 'amount', e.target.value)}
                />
                <input
                  className="field-input item-price"
                  type="number"
                  step="0.01"
                  min="0"
                  placeholder={t.form.item.price}
                  value={item.price}
                  onChange={e => setItem(i, 'price', e.target.value)}
                />
                <button type="button" className="remove-btn" onClick={() => removeItem(i)} title={t.form.remove}>✕</button>
              </div>
            ))}
            <button type="button" className="add-item-btn" onClick={addItem}>{t.form.addItem}</button>
          </div>
        </div>

        {/* splits */}
        <div className="field-group">
          <div className="splits-header">
            <label className="field-label">{t.form.splits}</label>
            <div className="splits-totals">
              <span className="splits-total-label">{t.form.splitsAllocated}: <strong>{allocated.toFixed(2)}</strong></span>
              <span className={`splits-remaining-label ${Math.abs(remaining) > 0.005 ? 'splits-remaining--warn' : ''}`}>
                {t.form.splitsRemaining}: <strong>{remaining.toFixed(2)}</strong>
              </span>
            </div>
          </div>

          <div className="splits-list">
            {receipt.splits.map((sp, i) => {
              const accsForClass = accountsForClass(sp.classId)
              return (
                <div key={i} className="split-row">
                  <select
                    className="field-input field-select split-class"
                    value={sp.classId}
                    onChange={e => setSplit(i, 'classId', e.target.value)}
                    disabled={dataLoading}
                  >
                    <option value="">{dataLoading ? '…' : t.form.selectClass}</option>
                    {classes.map(c => (
                      <option key={c.id} value={c.id}>{c.name}</option>
                    ))}
                  </select>

                  <select
                    className="field-input field-select split-account"
                    value={sp.accountId}
                    onChange={e => setSplit(i, 'accountId', e.target.value)}
                    disabled={dataLoading}
                    required
                  >
                    <option value="">{dataLoading ? '…' : t.form.selectAccount}</option>
                    {accsForClass.map(a => (
                      <option key={a.id} value={a.id}>[{a.shortcut}] {a.name}</option>
                    ))}
                  </select>

                  <input
                    className="field-input split-amount"
                    type="number"
                    step="0.01"
                    min="0"
                    placeholder={t.form.splitAmount}
                    value={sp.amount}
                    onChange={e => setSplit(i, 'amount', e.target.value)}
                    required
                  />

                  <button type="button" className="remove-btn" onClick={() => removeSplit(i)} title={t.form.removeSplit}>✕</button>
                </div>
              )
            })}

            <button type="button" className="add-item-btn" onClick={addSplit}>{t.form.addSplit}</button>
          </div>
        </div>

        {/* actions */}
        <div className="form-actions">
          <button
            type="submit"
            className="cta-button"
            disabled={state === 'sending' || !canSubmit}
          >
            {state === 'sending' ? (
              <><div className="spinner" />{t.form.sending}</>
            ) : (
              <>{t.form.send}</>
            )}
          </button>
          <button
            type="button"
            className="ghost-button"
            onClick={onCancel}
            disabled={state === 'sending'}
          >
            {t.form.cancel}
          </button>
        </div>
      </form>

      {receipt.fileId && (
        <div className="form-image-panel">
          <AuthImage
            src={`${FILE_SERVICE_URL}/files/${receipt.fileId}`}
            alt="receipt"
            className="form-receipt-image"
          />
        </div>
      )}
      </div>
    </div>
  )
}

// ── helpers ──────────────────────────────────────────────────────────────────

function normalise(r, userSub) {
  return {
    id: r?.id ?? '',
    fileId: r?.fileId ?? '',
    receiptOwnerId: r?.receiptOwnerId || userSub || '',
    accountId: r?.accountId ?? '',
    totalPrice: r?.totalPrice ?? 0,
    taxes: r?.taxes ?? 0,
    time: r?.time ?? new Date().toISOString(),
    items: (r?.items ?? []).map(it => ({
      name: it.name ?? '',
      amount: it.amount ?? 1,
      price: it.price ?? 0,
    })),
    splits: (r?.splits ?? [{ classId: '', accountId: '', amount: r?.totalPrice ?? '' }]).map(sp => ({
      classId: sp.classId ?? '',
      accountId: sp.accountId ?? '',
      amount: sp.amount ?? '',
    })),
  }
}

function toDatetimeLocal(iso) {
  try {
    const d = new Date(iso)
    return d.toISOString().slice(0, 16)
  } catch {
    return ''
  }
}
