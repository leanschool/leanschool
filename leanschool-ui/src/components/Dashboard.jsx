import { useTranslation } from '../i18n/useTranslation'
import { useAuth } from '../auth/useAuth'
import { hasFeature } from '../auth/permissions'
import './Dashboard.css'

export default function Dashboard({ onScan, onManageUsers, onManageClasses, onManageData, onManageTemplates, onTimetablePlanner }) {
  const { t } = useTranslation()
  const { user } = useAuth()

  const firstName = user?.given_name || user?.preferred_username || ''

  const cards = [
    hasFeature(user, 'scanReceipt') && {
      icon: '⬡',
      title: t.dashboard.scanCard.title,
      body:  t.dashboard.scanCard.body,
      onClick: onScan,
    },
    hasFeature(user, 'manageUsers') && {
      icon: '◉',
      title: t.dashboard.usersCard.title,
      body:  t.dashboard.usersCard.body,
      onClick: onManageUsers,
    },
    hasFeature(user, 'manageClasses') && {
      icon: '◈',
      title: t.dashboard.classesCard.title,
      body:  t.dashboard.classesCard.body,
      onClick: onManageClasses,
    },
    hasFeature(user, 'manageData') && {
      icon: '⬡',
      title: t.dashboard.dataCard.title,
      body:  t.dashboard.dataCard.body,
      onClick: onManageData,
    },
    hasFeature(user, 'manageTemplates') && {
      icon: '⬡',
      title: t.dashboard.templatesCard.title,
      body:  t.dashboard.templatesCard.body,
      onClick: onManageTemplates,
    },
    (hasFeature(user, 'timetablePlanner') || hasFeature(user, 'timetableView')) && {
      icon: '◈',
      title: t.dashboard.timetableCard.title,
      body:  t.dashboard.timetableCard.body,
      onClick: onTimetablePlanner,
    },
  ].filter(Boolean)

  return (
    <div className="dashboard">
      <div className="orb orb-1" />
      <div className="orb orb-2" />

      <nav className="dashboard-nav">
        <img src="/logo.svg" alt="leanschool" className="dashboard-nav-logo" />
      </nav>

      <section className="dashboard-hero">
        <h1 className="dashboard-title">
          {t.dashboard.greeting}{firstName && `, ${firstName}`}
        </h1>
        <p className="dashboard-subtitle">{t.dashboard.subtitle}</p>
      </section>

      <section className="dashboard-cards">
        {cards.length > 0 ? cards.map((card, i) => (
          <div key={i} className="dashboard-card" onClick={card.onClick}>
            <div className="dashboard-card-icon">{card.icon}</div>
            <h3>{card.title}</h3>
            <p>{card.body}</p>
            <span className="dashboard-card-arrow">→</span>
          </div>
        )) : (
          <p style={{ color: 'var(--text-3)', gridColumn: '1 / -1' }}>{t.dashboard.noAccess}</p>
        )}
      </section>
    </div>
  )
}
