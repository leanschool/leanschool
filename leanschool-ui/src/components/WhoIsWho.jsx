import { useState, useEffect } from 'react'
import { useTranslation } from '../i18n/useTranslation'
import { useAuth } from '../auth/useAuth'
import { config } from '../config'
import './WhoIsWho.css'

const API = config.leanschoolUrl

export default function WhoIsWho({ onBack }) {
  const { t } = useTranslation()
  const { authFetch } = useAuth()
  const [groups, setGroups] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)

  useEffect(() => {
    authFetch(`${API}/registration/who-is-who`)
      .then(r => r.ok ? r.json() : Promise.reject(r.status))
      .then(data => { setGroups(data); setLoading(false) })
      .catch(err => { setError(String(err)); setLoading(false) })
  }, []) // eslint-disable-line react-hooks/exhaustive-deps

  return (
    <div className="wiw-page">
      <div className="orb orb-1" />
      <div className="orb orb-2" />

      <nav className="wiw-nav">
        <button className="ghost-button wiw-back" onClick={onBack}>← {t.common?.back || 'Back'}</button>
        <h1 className="wiw-title">{t.whoIsWho.title}</h1>
      </nav>

      <p className="wiw-subtitle">{t.whoIsWho.subtitle}</p>

      {loading && <div className="wiw-loading">…</div>}
      {error && <div className="wiw-error">{error}</div>}

      <div className="wiw-grid">
        {groups.map(group => (
          <div key={group.name} className="wiw-card">
            <h2 className="wiw-role-name">
              {t.registration?.roles?.[group.name] || group.name}
            </h2>
            {group.members.length === 0 ? (
              <p className="wiw-empty">{t.whoIsWho.noMembers}</p>
            ) : (
              <ul className="wiw-members">
                {group.members.map(m => (
                  <li key={m.id} className="wiw-member">
                    {m.firstName || m.lastName
                      ? `${m.firstName} ${m.lastName}`.trim()
                      : m.username}
                  </li>
                ))}
              </ul>
            )}
          </div>
        ))}
      </div>
    </div>
  )
}
