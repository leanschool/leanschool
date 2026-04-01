import { useState, useEffect } from 'react'
import { useAuth } from '../../auth/useAuth'
import { useTranslation } from '../../i18n/useTranslation'
import './domain-components.css'

import { config } from '../../config'

const API = config.leanschoolUrl

export default function AccountView({ id }) {
  const { t } = useTranslation()
  const { authFetch } = useAuth()
  const [account, setAccount] = useState(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)

  useEffect(() => {
    if (!id) return
    setLoading(true)
    setError(null)
    authFetch(`${API}/accounts/${id}`)
      .then(res => {
        if (!res.ok) throw new Error(`Failed to load account (${res.status})`)
        return res.json()
      })
      .then(data => setAccount(data))
      .catch(err => setError(err.message))
      .finally(() => setLoading(false))
  }, [id]) // eslint-disable-line react-hooks/exhaustive-deps

  if (loading) {
    return (
      <div className="dv-loading">
        <div className="spinner" />
      </div>
    )
  }

  if (error) return <div className="dv-error">{error}</div>
  if (!account) return null

  const fmt = val => val != null ? Number(val).toFixed(2) : '—'
  const fmtDate = val => val ? new Date(val).toLocaleDateString() : '—'

  return (
    <div className="dv-card">
      <div className="dv-row">
        <span className="dv-label">{t.domain.common.id}</span>
        <span className="dv-value">{account.id}</span>
      </div>
      <div className="dv-row">
        <span className="dv-label">{t.domain.account.name}</span>
        <span className="dv-value">{account.name}</span>
      </div>
      <div className="dv-row">
        <span className="dv-label">{t.domain.account.shortcut}</span>
        <span className="dv-value">{account.shortcut}</span>
      </div>
      <div className="dv-row">
        <span className="dv-label">{t.domain.account.budget}</span>
        <span className="dv-value">{fmt(account.budget)}</span>
      </div>
      <div className="dv-row">
        <span className="dv-label">{t.domain.account.spent}</span>
        <span className="dv-value">{fmt(account.spent)}</span>
      </div>
      <div className="dv-row">
        <span className="dv-label">{t.domain.account.balance}</span>
        <span className="dv-value">{fmt(account.balance)}</span>
      </div>
      {account.classId && (
        <div className="dv-row">
          <span className="dv-label">{t.domain.account.class}</span>
          <span className="dv-value">{account.classId}</span>
        </div>
      )}
      {account.validFrom && (
        <div className="dv-row">
          <span className="dv-label">{t.domain.account.validFrom}</span>
          <span className="dv-value">{fmtDate(account.validFrom)}</span>
        </div>
      )}
      {account.validTo && (
        <div className="dv-row">
          <span className="dv-label">{t.domain.account.validTo}</span>
          <span className="dv-value">{fmtDate(account.validTo)}</span>
        </div>
      )}
    </div>
  )
}
