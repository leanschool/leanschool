import { useEffect, useState } from 'react'
import { useTranslation } from '../i18n/useTranslation'
import { useAuth } from '../auth/useAuth'
import { config } from '../config'
import './UserManagement.css'

const API = config.leanschoolUrl

export default function UserManagement({ onBack }) {
  const { t } = useTranslation()
  const { authFetch } = useAuth()
  const [requests, setRequests] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [acting, setActing] = useState(null)

  useEffect(() => {
    authFetch(`${API}/users/registrations`)
      .then(res => {
        if (!res.ok) throw new Error(`${res.status}`)
        return res.json()
      })
      .then(data => setRequests(data))
      .catch(err => setError(err.message))
      .finally(() => setLoading(false))
  }, [authFetch])

  async function act(id, action) {
    setActing(id + action)
    try {
      const res = await authFetch(`${API}/users/registrations/${id}/${action}`, { method: 'POST' })
      if (!res.ok) throw new Error(await res.text())
      setRequests(prev => prev.map(r => r.id === id ? { ...r, status: action === 'approve' ? 'approved' : 'denied' } : r))
    } catch (err) {
      alert(err.message)
    } finally {
      setActing(null)
    }
  }

  function statusLabel(status) {
    if (status === 'pending')  return <span className="um-badge um-badge-pending">{t.userManagement.statusPending}</span>
    if (status === 'approved') return <span className="um-badge um-badge-approved">{t.userManagement.statusApproved}</span>
    if (status === 'denied')   return <span className="um-badge um-badge-denied">{t.userManagement.statusDenied}</span>
    return <span className="um-badge">{status}</span>
  }

  return (
    <div className="um-page">
      <div className="orb orb-1" />
      <div className="orb orb-2" />

      <div className="um-container">
        <div className="um-header">
          <div>
            <h2 className="um-title">{t.userManagement.title}</h2>
            <p className="um-subtitle">{t.userManagement.subtitle}</p>
          </div>
        </div>

        {loading && <p className="um-info">…</p>}
        {error && <p className="um-error">{error}</p>}

        {!loading && !error && requests.length === 0 && (
          <p className="um-info">{t.userManagement.noRequests}</p>
        )}

        {requests.length > 0 && (
          <div className="um-table-wrap">
            <table className="um-table">
              <thead>
                <tr>
                  <th>Sub</th>
                  <th>Email</th>
                  <th>Roles</th>
                  <th>Status</th>
                  <th></th>
                </tr>
              </thead>
              <tbody>
                {requests.map(req => (
                  <tr key={req.id}>
                    <td className="um-sub">{req.userSub}</td>
                    <td>{req.email}</td>
                    <td>{req.desiredRoles?.join(', ')}</td>
                    <td>{statusLabel(req.status)}</td>
                    <td className="um-actions">
                      {req.status === 'pending' && (
                        <>
                          <button
                            className="um-btn um-btn-approve"
                            disabled={acting !== null}
                            onClick={() => act(req.id, 'approve')}
                          >
                            {t.userManagement.approve}
                          </button>
                          <button
                            className="um-btn um-btn-deny"
                            disabled={acting !== null}
                            onClick={() => act(req.id, 'deny')}
                          >
                            {t.userManagement.deny}
                          </button>
                        </>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
      <button className="page-back-btn" onClick={onBack}>← {t.userManagement.back}</button>
    </div>
  )
}
