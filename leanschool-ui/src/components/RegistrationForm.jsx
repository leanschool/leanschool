import { useState, useEffect } from 'react'
import { useTranslation } from '../i18n/useTranslation'
import { useAuth } from '../auth/useAuth'
import { config } from '../config'
import './RegistrationForm.css'

const API = config.leanschoolUrl

const TEACHER_ROLE_NAME = 'teacher'

export default function RegistrationForm({ onRegistered }) {
  const { t } = useTranslation()
  const { authFetch, user } = useAuth()
  const [selectedRoles, setSelectedRoles] = useState([])
  const [email, setEmail] = useState(user?.email ?? '')
  const [classIds, setClassIds] = useState([])
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState(null)
  const [roleOptions, setRoleOptions] = useState([])
  const [rolesLoading, setRolesLoading] = useState(true)
  const [allClasses, setAllClasses] = useState([])

  useEffect(() => {
    authFetch(`${API}/users/role-mappings`)
      .then(r => r.ok ? r.json() : [])
      .then(data => { setRoleOptions(data); setRolesLoading(false) })
      .catch(() => setRolesLoading(false))
  }, []) // eslint-disable-line react-hooks/exhaustive-deps

  // Load classes when teacher role is among selected roles
  useEffect(() => {
    if (selectedRoles.includes(TEACHER_ROLE_NAME)) {
      authFetch(`${API}/registration/school-classes`)
        .then(r => r.ok ? r.json() : [])
        .then(data => setAllClasses(data ?? []))
        .catch(() => {})
    } else {
      setAllClasses([])
      setClassIds([])
    }
  }, [selectedRoles]) // eslint-disable-line react-hooks/exhaustive-deps

  const toggleRole = name => {
    setSelectedRoles(prev =>
      prev.includes(name) ? prev.filter(r => r !== name) : [...prev, name]
    )
  }

  const toggleClass = id => {
    setClassIds(prev =>
      prev.includes(id) ? prev.filter(c => c !== id) : [...prev, id]
    )
  }

  async function handleSubmit(e) {
    e.preventDefault()
    if (selectedRoles.length === 0) return
    setSubmitting(true)
    setError(null)
    try {
      const res = await authFetch(`${API}/users/register`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ desiredRoles: selectedRoles, email, classIds }),
      })
      if (!res.ok) throw new Error(await res.text())
      onRegistered()
    } catch (err) {
      setError(err.message)
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="reg-page">
      <div className="orb orb-1" />
      <div className="orb orb-2" />
      <div className="reg-card">
        <h2 className="reg-title">{t.registration.title}</h2>
        <p className="reg-subtitle">{t.registration.subtitle}</p>

        <form onSubmit={handleSubmit} className="reg-form">
          <fieldset className="reg-fieldset">
            <legend className="reg-legend">{t.registration.roleLabel}</legend>
            {rolesLoading ? (
              <span className="reg-hint">…</span>
            ) : (
              roleOptions.map(opt => (
                <label key={opt.name} className="reg-radio-label">
                  <input
                    type="checkbox"
                    value={opt.name}
                    checked={selectedRoles.includes(opt.name)}
                    onChange={() => toggleRole(opt.name)}
                  />
                  {t.registration.roles?.[opt.name] || opt.description || opt.name}
                </label>
              ))
            )}
          </fieldset>

          <label className="reg-field-label">
            {t.registration.emailLabel}
            <input
              type="email"
              className="reg-input"
              value={email}
              onChange={e => setEmail(e.target.value)}
              required
            />
          </label>

          {selectedRoles.includes(TEACHER_ROLE_NAME) && allClasses.length > 0 && (
            <fieldset className="reg-fieldset">
              <legend className="reg-legend">{t.registration.classesLabel}</legend>
              {allClasses.map(c => (
                <label key={c.id} className="reg-radio-label">
                  <input
                    type="checkbox"
                    checked={classIds.includes(c.id)}
                    onChange={() => toggleClass(c.id)}
                  />
                  {c.name}
                </label>
              ))}
            </fieldset>
          )}

          {error && <p className="reg-error">{error}</p>}

          <button className="cta-button" type="submit" disabled={submitting || selectedRoles.length === 0}>
            {submitting ? '…' : t.registration.submit}
          </button>
        </form>
      </div>
    </div>
  )
}
