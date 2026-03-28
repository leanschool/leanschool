import { useTranslation } from '../i18n/useTranslation'
import './UserInfo.css'

export default function UserInfo({ user, onClose, onLogout, onEditProfile }) {
  const { t } = useTranslation()
  const initials = [user?.given_name, user?.family_name]
    .filter(Boolean)
    .map(n => n[0].toUpperCase())
    .join('') || user?.preferred_username?.[0]?.toUpperCase() || '?'

  const roles = user?.realm_access?.roles?.filter(r => !['offline_access', 'uma_authorization'].includes(r)) ?? []

  return (
    <div className="userinfo-overlay" onClick={onClose}>
      <div className="userinfo-panel" onClick={e => e.stopPropagation()}>
        <button className="userinfo-close" onClick={onClose} aria-label="Close">✕</button>

        <div className="userinfo-avatar">{initials}</div>

        <div className="userinfo-name">
          {[user?.given_name, user?.family_name].filter(Boolean).join(' ') || user?.preferred_username}
        </div>
        {user?.email && (
          <div className="userinfo-email">{user.email}</div>
        )}

        {roles.length > 0 && (
          <div className="userinfo-roles">
            <div className="userinfo-roles-label">Roles</div>
            <div className="userinfo-roles-list">
              {roles.map(r => (
                <span key={r} className="userinfo-role-badge">{r}</span>
              ))}
            </div>
          </div>
        )}

        <button className="userinfo-edit-btn" onClick={onEditProfile}>
          {t.userInfo.editProfile}
        </button>

        <button className="userinfo-logout-btn" onClick={onLogout}>
          {t.userInfo.signOut}
        </button>
      </div>
    </div>
  )
}
