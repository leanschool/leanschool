import { useState, useEffect } from 'react'
import { useTranslation } from '../i18n/useTranslation'
import { useAuth } from '../auth/useAuth'
import { getUserRoles } from '../auth/permissions'
import { config } from '../config'
import './WelcomePage.css'

const API = config.leanschoolUrl

export default function WelcomePage({ onDone }) {
  const { t } = useTranslation()
  const { authFetch, user } = useAuth()
  const [iban, setIban] = useState('')
  const [address, setAddress] = useState('')
  const [phone, setPhone] = useState('')
  const [classIds, setClassIds] = useState([])
  const [allClasses, setAllClasses] = useState([])
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState(null)

  const isTeacher = getUserRoles(user).includes('teacher')

  // Load existing profile and school classes on mount
  useEffect(() => {
    authFetch(`${API}/users/profile`)
      .then(r => r.ok ? r.json() : null)
      .then(profile => {
        if (profile) {
          setIban(profile.iban ?? '')
          setAddress(profile.address ?? '')
          setPhone(profile.phone ?? '')
          setClassIds(profile.classIds ?? [])
        }
      })
      .catch(() => {})

    if (isTeacher) {
      authFetch(`${API}/school-classes`)
        .then(r => r.ok ? r.json() : [])
        .then(data => setAllClasses(data ?? []))
        .catch(() => {})
    }
  }, []) // eslint-disable-line react-hooks/exhaustive-deps

  const toggleClass = id => {
    setClassIds(prev =>
      prev.includes(id) ? prev.filter(c => c !== id) : [...prev, id]
    )
  }

  async function save(profileSkipped) {
    setSaving(true)
    setError(null)
    try {
      const res = await authFetch(`${API}/users/profile`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ iban, address, phone, classIds, profileSkipped }),
      })
      if (!res.ok) throw new Error(await res.text())
      onDone()
    } catch (err) {
      setError(err.message)
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="welcome-page">
      <div className="orb orb-1" />
      <div className="orb orb-2" />
      <div className="welcome-card">
        <h2 className="welcome-title">{t.welcome.title}</h2>
        <p className="welcome-subtitle">{t.welcome.subtitle}</p>

        <form onSubmit={e => { e.preventDefault(); save(false) }} className="welcome-form">
          <label className="welcome-field-label">
            {t.welcome.ibanLabel}
            <input className="welcome-input" type="text" value={iban} onChange={e => setIban(e.target.value)} />
          </label>

          <label className="welcome-field-label">
            {t.welcome.addressLabel}
            <input className="welcome-input" type="text" value={address} onChange={e => setAddress(e.target.value)} />
          </label>

          <label className="welcome-field-label">
            {t.welcome.phoneLabel}
            <input className="welcome-input" type="tel" value={phone} onChange={e => setPhone(e.target.value)} />
          </label>

          {isTeacher && allClasses.length > 0 && (
            <div className="welcome-field-label">
              <span>{t.welcome.classesLabel}</span>
              <div className="welcome-classes">
                {allClasses.map(c => (
                  <label key={c.id} className="welcome-class-option">
                    <input
                      type="checkbox"
                      checked={classIds.includes(c.id)}
                      onChange={() => toggleClass(c.id)}
                    />
                    {c.name}
                  </label>
                ))}
              </div>
            </div>
          )}

          {error && <p className="welcome-error">{error}</p>}

          <button className="cta-button" type="submit" disabled={saving}>
            {saving ? '…' : t.welcome.save}
          </button>

          <button type="button" className="ghost-button welcome-skip" onClick={() => save(true)} disabled={saving}>
            {t.welcome.skip}
          </button>
        </form>
      </div>
    </div>
  )
}
