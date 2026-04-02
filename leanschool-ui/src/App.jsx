import { useState, useEffect } from 'react'
import { I18nProvider, useTranslation } from './i18n/useTranslation'
import { AuthProvider, useAuth } from './auth/useAuth'
import { getUserRoles } from './auth/permissions'
import LandingPage from './components/LandingPage'
import Dashboard from './components/Dashboard'
import ScanReceipt from './components/ScanReceipt'
import ReceiptForm from './components/ReceiptForm'
import SchoolClassManager from './components/SchoolClassManager'
import UserInfo from './components/UserInfo'
import LangDropdown from './components/LangDropdown'
import ThemeToggle from './components/ThemeToggle'
import RegistrationWizard from './components/RegistrationWizard'
import AdminRegistrationDashboard from './components/AdminRegistrationDashboard'
import AwaitingApproval from './components/AwaitingApproval'
import DeniedScreen from './components/DeniedScreen'
import WelcomePage from './components/WelcomePage'
import DataDashboard from './components/DataDashboard'
import TemplateManager from './components/TemplateManager'
import TimetablePlanner from './components/timetable-planner/TimetablePlanner'
import WhoIsWho from './components/WhoIsWho'
import './App.css'
import './components/UserInfo.css'
import './components/WhoIsWho.css'
import './components/RegistrationWizard.css'
import './components/AdminRegistrationDashboard.css'
import { config } from './config'

const API = config.leanschoolUrl

const BUSINESS_ROLES = ['teacher', 'school-management', 'user_management']

function AppContent() {
  const { tokens, user, loading, login, logout, authFetch } = useAuth()
  const { t } = useTranslation()
  const [view, setView] = useState('dashboard')
  const [extracted, setExtracted] = useState(null)
  const [showUserInfo, setShowUserInfo] = useState(false)

  // User status fetched from backend for authenticated users.
  const [userStatus, setUserStatus] = useState(null)
  // Start true when tokens exist so no dashboard flash before /users/me resolves.
  const [statusLoading, setStatusLoading] = useState(!!tokens)
  const [statusError, setStatusError] = useState(false)
  const [retryCount, setRetryCount] = useState(0)

  // Fetch /users/me whenever the user becomes authenticated.
  useEffect(() => {
    if (!tokens || !user) {
      setStatusLoading(false)
      return
    }
    setStatusLoading(true)
    setStatusError(false)
    authFetch(`${API}/users/me`)
      .then(res => {
        if (!res.ok) throw new Error(`users/me ${res.status}`)
        return res.json()
      })
      .then(data => setUserStatus(data))
      .catch(() => setStatusError(true))
      .finally(() => setStatusLoading(false))
  }, [tokens, user?.sub, retryCount]) // eslint-disable-line react-hooks/exhaustive-deps

  // Show a minimal loading screen while the auth callback is being processed.
  if (loading || (tokens && statusLoading)) {
    return (
      <div style={{ minHeight: '100svh', background: 'var(--bg-page)', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
        <div style={{ color: 'var(--text-3)', fontSize: 15 }}>Loading…</div>
        <LangDropdown solo />
        <ThemeToggle solo />
      </div>
    )
  }

  // API error fetching user status — show retry screen (never fall through to Dashboard).
  if (tokens && statusError) {
    return (
      <div style={{ minHeight: '100svh', background: 'var(--bg-page)', display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', gap: 16 }}>
        <div style={{ color: 'var(--text-3)', fontSize: 15 }}>Unable to reach server.</div>
        <button className="ghost-button" onClick={() => setRetryCount(c => c + 1)}>Retry</button>
        <LangDropdown solo />
        <ThemeToggle solo />
      </div>
    )
  }

  // Not authenticated → marketing landing page with sign-in button.
  if (!tokens) {
    return (
      <>
        <LandingPage onLogin={login} />
        <LangDropdown solo />
        <ThemeToggle solo />
      </>
    )
  }

  // ── Authenticated: check business roles and registration status ──────────────

  const roles = getUserRoles(user)
  const hasBusinessRole = BUSINESS_ROLES.some(r => roles.includes(r))

  if (!hasBusinessRole) {
    // Non-business users must always have a valid userStatus — never fall through to Dashboard.
    if (!userStatus) return null
    const rs = userStatus.registrationStatus
    if (rs === 'none') {
      return (
        <>
          <RegistrationWizard onRegistered={() => setUserStatus({ ...userStatus, registrationStatus: 'pending' })} />
          <LangDropdown solo />
          <ThemeToggle solo />
        </>
      )
    }
    if (rs === 'pending') {
      return (
        <>
          <AwaitingApproval />
          <LangDropdown solo />
          <ThemeToggle solo />
        </>
      )
    }
    if (rs === 'denied') {
      return (
        <>
          <DeniedScreen />
          <LangDropdown solo />
          <ThemeToggle solo />
        </>
      )
    }
  }

  // Has business roles — check welcome page condition.
  if (
    hasBusinessRole &&
    userStatus &&
    userStatus.registrationStatus === 'approved' &&
    !userStatus.profileComplete &&
    !userStatus.profileSkipped
  ) {
    return (
      <>
        <WelcomePage onDone={() => setUserStatus({ ...userStatus, profileComplete: true })} />
        <LangDropdown solo />
        <ThemeToggle solo />
      </>
    )
  }

  // ── Authenticated views ────────────────────────────────────────────────────

  const initials = [user?.given_name, user?.family_name]
    .filter(Boolean)
    .map(n => n[0].toUpperCase())
    .join('') || user?.preferred_username?.[0]?.toUpperCase() || '?'

  let content

  if (view === 'scan') {
    content = (
      <ScanReceipt
        onExtracted={receipt => { setExtracted(receipt); setView('review') }}
        onCancel={() => setView('dashboard')}
      />
    )
  } else if (view === 'review') {
    content = (
      <ReceiptForm
        receipt={extracted}
        onSaved={() => setView('success')}
        onCancel={() => setView('scan')}
      />
    )
  } else if (view === 'success') {
    content = <SuccessScreen onReset={() => setView('scan')} onBack={() => setView('dashboard')} />
  } else if (view === 'users') {
    content = <AdminRegistrationDashboard onBack={() => setView('dashboard')} />
  } else if (view === 'classes') {
    content = <SchoolClassManager onBack={() => setView('dashboard')} />
  } else if (view === 'profile') {
    content = <WelcomePage onDone={() => setView('dashboard')} />
  } else if (view === 'data') {
    content = <DataDashboard onBack={() => setView('dashboard')} />
  } else if (view === 'templates') {
    content = <TemplateManager onBack={() => setView('dashboard')} />
  } else if (view === 'who-is-who') {
    content = <WhoIsWho onBack={() => setView('dashboard')} />
  } else if (view === 'timetable-planner') {
    content = <TimetablePlanner onBack={() => setView('dashboard')} />
  } else {
    content = (
      <Dashboard
        onScan={() => setView('scan')}
        onManageUsers={() => setView('users')}
        onManageClasses={() => setView('classes')}
        onManageData={() => setView('data')}
        onManageTemplates={() => setView('templates')}
        onTimetablePlanner={() => setView('timetable-planner')}
      />
    )
  }

  return (
    <>
      {content}

      <LangDropdown />
      <ThemeToggle />

      {/* Fixed who-is-who button — visible on all authenticated views */}
      <button className="who-icon-btn" onClick={() => setView('who-is-who')} title={t.whoIsWho?.title}>
        ◎
      </button>

      {/* Fixed user icon — visible on all authenticated views */}
      <button className="user-icon-btn" onClick={() => setShowUserInfo(true)} title={user?.preferred_username}>
        {initials}
      </button>

      {showUserInfo && (
        <UserInfo
          user={user}
          onClose={() => setShowUserInfo(false)}
          onLogout={logout}
          onEditProfile={() => { setShowUserInfo(false); setView('profile') }}
        />
      )}
    </>
  )
}

function SuccessScreen({ onReset, onBack }) {
  const { t } = useTranslation()
  return (
    <div className="success-page">
      <div className="orb orb-1" />
      <div className="orb orb-2" />
      <div className="success-card">
        <div className="success-icon">✦</div>
        <h2 className="success-title">{t.success.title}</h2>
        <p className="success-subtitle">{t.success.subtitle}</p>
        <button className="cta-button" onClick={onReset}>
          {t.success.scanAnother}
        </button>
        <button className="ghost-button" onClick={onBack}>
          {t.success.backToDashboard}
        </button>
      </div>
    </div>
  )
}

export default function App() {
  return (
    <I18nProvider>
      <AuthProvider>
        <AppContent />
      </AuthProvider>
    </I18nProvider>
  )
}
