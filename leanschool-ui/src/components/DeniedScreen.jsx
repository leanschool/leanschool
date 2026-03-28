import { useTranslation } from '../i18n/useTranslation'
import { useAuth } from '../auth/useAuth'
import './RegistrationForm.css'

export default function DeniedScreen() {
  const { t } = useTranslation()
  const { logout } = useAuth()

  return (
    <div className="reg-page">
      <div className="orb orb-1" />
      <div className="orb orb-2" />
      <div className="reg-card reg-card--center">
        <div className="reg-status-icon">✕</div>
        <h2 className="reg-title">{t.denied.title}</h2>
        <p className="reg-subtitle">{t.denied.subtitle}</p>
        <button className="ghost-button" onClick={logout}>
          {t.common.back}
        </button>
      </div>
    </div>
  )
}
