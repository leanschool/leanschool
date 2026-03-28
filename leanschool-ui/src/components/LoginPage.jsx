import { useTranslation } from '../i18n/useTranslation'
import './LoginPage.css'

export default function LoginPage({ onLogin }) {
  const { t } = useTranslation()
  return (
    <div className="login-page">
      <div className="orb orb-1" />
      <div className="orb orb-2" />
      <div className="orb orb-3" />

      <div className="login-card">
        <img src="/logo.svg" alt="LeanSchool" className="login-logo" />

        <button className="login-btn" onClick={onLogin}>
          <span className="login-btn-icon">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <path d="M15 3h4a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2h-4" />
              <polyline points="10 17 15 12 10 7" />
              <line x1="15" y1="12" x2="3" y2="12" />
            </svg>
          </span>
          {t.login.signIn}
        </button>
      </div>
    </div>
  )
}
