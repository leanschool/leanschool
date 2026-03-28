import { useState, useEffect } from 'react'
import { useTranslation } from '../i18n/useTranslation'
import { useAuth } from '../auth/useAuth'
import './RegistrationWizard.css'

const API = import.meta.env.VITE_API_URL || 'http://localhost:8080'

const STEPS = {
  ROLE_SELECTION: 1,
  PERSONAL_INFO: 2,
  ROLE_DETAILS: 3,
  CONFIRMATION: 4
}

export default function RegistrationWizard({ onRegistered }) {
  const { t } = useTranslation()
  const { authFetch, user } = useAuth()
  const [currentStep, setCurrentStep] = useState(STEPS.ROLE_SELECTION)
  const [selectedRoles, setSelectedRoles] = useState([])
  const [email, setEmail] = useState(user?.email ?? '')
  const [firstName, setFirstName] = useState('')
  const [lastName, setLastName] = useState('')
  const [classIds, setClassIds] = useState([])
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState(null)
  const [roleOptions, setRoleOptions] = useState([])
  const [rolesLoading, setRolesLoading] = useState(true)
  const [allClasses, setAllClasses] = useState([])
  const [classesLoading, setClassesLoading] = useState(false)

  useEffect(() => {
    authFetch(`${API}/users/role-mappings`)
      .then(r => r.ok ? r.json() : [])
      .then(data => { setRoleOptions(data); setRolesLoading(false) })
      .catch(() => setRolesLoading(false))
  }, []) // eslint-disable-line react-hooks/exhaustive-deps

  useEffect(() => {
    if (selectedRoles.includes('teacher') && currentStep >= STEPS.ROLE_DETAILS) {
      setClassesLoading(true)
      authFetch(`${API}/registration/school-classes`)
        .then(r => r.ok ? r.json() : [])
        .then(data => setAllClasses(data ?? []))
        .catch(() => {})
        .finally(() => setClassesLoading(false))
    }
  }, [selectedRoles, currentStep]) // eslint-disable-line react-hooks/exhaustive-deps

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

  const nextStep = () => {
    if (currentStep === STEPS.ROLE_SELECTION && selectedRoles.length === 0) return
    if (currentStep === STEPS.PERSONAL_INFO && !firstName.trim() && !lastName.trim()) return
    setCurrentStep(prev => prev + 1)
  }

  const prevStep = () => {
    setCurrentStep(prev => prev - 1)
  }

  async function handleSubmit() {
    setSubmitting(true)
    setError(null)
    try {
      const registrationData = {
        desiredRoles: selectedRoles,
        personData: {
          name: lastName,
          prename: firstName
        },
        contactEmail: email
      }

      // Add teacher data if teacher role is selected
      if (selectedRoles.includes('teacher')) {
        registrationData.teacherData = {
          classIds: classIds
        }
      }

      const res = await authFetch(`${API}/registration/start`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(registrationData),
      })

      if (!res.ok) throw new Error(await res.text())
      onRegistered()
    } catch (err) {
      setError(err.message)
    } finally {
      setSubmitting(false)
    }
  }

  const renderStep = () => {
    switch (currentStep) {
      case STEPS.ROLE_SELECTION:
        return (
          <div className="rw-step">
            <h3 className="rw-step-title">{t.registration.roleLabel}</h3>
            <p className="rw-step-subtitle">{t.registration.subtitle}</p>
            
            {rolesLoading ? (
              <span className="rw-hint">…</span>
            ) : (
              <div className="rw-role-grid">
                {roleOptions.map(opt => (
                  <label key={opt.name} className="rw-role-card">
                    <input
                      type="checkbox"
                      value={opt.name}
                      checked={selectedRoles.includes(opt.name)}
                      onChange={() => toggleRole(opt.name)}
                      className="rw-role-checkbox"
                    />
                    <div className="rw-role-content">
                      <span className="rw-role-name">
                        {t.registration.roles?.[opt.name] || opt.description || opt.name}
                      </span>
                    </div>
                  </label>
                ))}
              </div>
            )}
          </div>
        )

      case STEPS.PERSONAL_INFO:
        return (
          <div className="rw-step">
            <h3 className="rw-step-title">{t.registration.personalInfoTitle}</h3>
            
            <label className="rw-field-label">
              {t.registration.firstNameLabel}
              <input
                type="text"
                className="rw-input"
                value={firstName}
                onChange={e => setFirstName(e.target.value)}
                required
              />
            </label>

            <label className="rw-field-label">
              {t.registration.lastNameLabel}
              <input
                type="text"
                className="rw-input"
                value={lastName}
                onChange={e => setLastName(e.target.value)}
                required
              />
            </label>

            <label className="rw-field-label">
              {t.registration.emailLabel}
              <input
                type="email"
                className="rw-input"
                value={email}
                onChange={e => setEmail(e.target.value)}
                required
              />
            </label>
          </div>
        )

      case STEPS.ROLE_DETAILS:
        return (
          <div className="rw-step">
            <h3 className="rw-step-title">{t.registration.roleDetailsTitle}</h3>
            
            {selectedRoles.includes('teacher') && (
              <div className="rw-section">
                <h4 className="rw-section-title">{t.registration.classesLabel}</h4>
                {classesLoading ? (
                  <span className="rw-hint">…</span>
                ) : allClasses.length > 0 ? (
                  <div className="rw-class-grid">
                    {allClasses.map(c => (
                      <label key={c.id} className="rw-class-card">
                        <input
                          type="checkbox"
                          checked={classIds.includes(c.id)}
                          onChange={() => toggleClass(c.id)}
                          className="rw-class-checkbox"
                        />
                        <span className="rw-class-name">{c.name}</span>
                      </label>
                    ))}
                  </div>
                ) : (
                  <p className="rw-info">{t.registration.noClassesAvailable}</p>
                )}
              </div>
            )}
          </div>
        )

      case STEPS.CONFIRMATION:
        return (
          <div className="rw-step">
            <h3 className="rw-step-title">{t.registration.confirmationTitle}</h3>
            
            <div className="rw-summary">
              <div className="rw-summary-section">
                <h4 className="rw-summary-title">{t.registration.personalInfoTitle}</h4>
                <p><strong>{t.registration.firstNameLabel}:</strong> {firstName}</p>
                <p><strong>{t.registration.lastNameLabel}:</strong> {lastName}</p>
                <p><strong>{t.registration.emailLabel}:</strong> {email}</p>
              </div>

              <div className="rw-summary-section">
                <h4 className="rw-summary-title">{t.registration.roleLabel}</h4>
                <p>{selectedRoles.map(r => t.registration.roles?.[r] || r).join(', ')}</p>
              </div>

              {selectedRoles.includes('teacher') && classIds.length > 0 && (
                <div className="rw-summary-section">
                  <h4 className="rw-summary-title">{t.registration.classesLabel}</h4>
                  <p>{allClasses.filter(c => classIds.includes(c.id)).map(c => c.name).join(', ')}</p>
                </div>
              )}
            </div>

            <p className="rw-confirmation-note">{t.registration.confirmationNote}</p>
          </div>
        )

      default:
        return null
    }
  }

  return (
    <div className="rw-page">
      <div className="orb orb-1" />
      <div className="orb orb-2" />
      
      <div className="rw-container">
        <div className="rw-header">
          <h2 className="rw-title">{t.registration.title}</h2>
          <div className="rw-progress">
            {Object.values(STEPS).map(step => (
              <div key={step} className={`rw-progress-step ${currentStep >= step ? 'rw-progress-step-active' : ''}`} />
            ))}
          </div>
        </div>

        {renderStep()}

        {error && <p className="rw-error">{error}</p>}

        <div className="rw-actions">
          {currentStep > STEPS.ROLE_SELECTION && (
            <button className="rw-btn rw-btn-secondary" onClick={prevStep} disabled={submitting}>
              {t.common.back}
            </button>
          )}

          {currentStep < STEPS.CONFIRMATION ? (
            <button className="rw-btn rw-btn-primary" onClick={nextStep} disabled={submitting}>
              {t.registration.next}
            </button>
          ) : (
            <button className="rw-btn rw-btn-primary" onClick={handleSubmit} disabled={submitting}>
              {submitting ? '…' : t.registration.submit}
            </button>
          )}
        </div>
      </div>
    </div>
  )
}