import { useState, useEffect, useRef } from 'react'
import { useTranslation } from '../i18n/useTranslation'
import { useAuth } from '../auth/useAuth'
import { config } from '../config'
import './RegistrationForm.css'

const API = config.leanschoolUrl
const POLL_INTERVAL = 15000 // 15 seconds

export default function AwaitingApproval({ onStatusChange }) {
  const { t } = useTranslation()
  const { logout, authFetch } = useAuth()
  const [polling] = useState(true)
  const intervalRef = useRef(null)

  useEffect(() => {
    if (!polling) return

    const checkStatus = async () => {
      try {
        const res = await authFetch(`${API}/users/me`)
        if (!res.ok) return
        const data = await res.json()
        if (data.registrationStatus !== 'pending') {
          onStatusChange?.(data)
        }
      } catch {
        // Ignore network errors, will retry on next interval
      }
    }

    // Check immediately on mount
    checkStatus()
    intervalRef.current = setInterval(checkStatus, POLL_INTERVAL)

    return () => {
      if (intervalRef.current) clearInterval(intervalRef.current)
    }
  }, [polling]) // eslint-disable-line react-hooks/exhaustive-deps

  return (
    <div className="reg-page">
      <div className="orb orb-1" />
      <div className="orb orb-2" />
      <div className="reg-card reg-card--center">
        <div className="reg-status-icon">&#x23F3;</div>
        <h2 className="reg-title">{t.awaiting.title}</h2>
        <p className="reg-subtitle">{t.awaiting.subtitle}</p>
        <p className="reg-hint">{t.awaiting.polling}</p>
        <button className="ghost-button" onClick={logout}>
          {t.common.back}
        </button>
      </div>
    </div>
  )
}
