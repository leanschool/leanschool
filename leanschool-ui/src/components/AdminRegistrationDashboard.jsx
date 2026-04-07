import { useEffect, useState } from 'react'
import { useTranslation } from '../i18n/useTranslation'
import { useAuth } from '../auth/useAuth'
import { config } from '../config'
import './AdminRegistrationDashboard.css'

const API = config.leanschoolUrl

export default function AdminRegistrationDashboard({ onBack }) {
  const { t } = useTranslation()
  const { authFetch } = useAuth()
  const [requests, setRequests] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [acting, setActing] = useState(null)
  const [filterStatus, setFilterStatus] = useState('all')
  const [searchTerm, setSearchTerm] = useState('')

  useEffect(() => {
    fetchRequests()
  }, [authFetch]) // eslint-disable-line react-hooks/exhaustive-deps

  const fetchRequests = () => {
    setLoading(true)
    authFetch(`${API}/registration/workflow`)
      .then(res => {
        if (!res.ok) throw new Error(`${res.status}`)
        return res.json()
      })
      .then(data => setRequests(data))
      .catch(err => setError(err.message))
      .finally(() => setLoading(false))
  }

  async function act(id, action) {
    setActing(id + action)
    try {
      const res = await authFetch(`${API}/registration/${id}/${action}`, { method: 'POST' })
      if (!res.ok) throw new Error(await res.text())
      fetchRequests() // Refresh the list
    } catch (err) {
      alert(err.message)
    } finally {
      setActing(null)
    }
  }

  function statusLabel(status) {
    if (status === 'pending')  return <span className="ard-badge ard-badge-pending">{t.userManagement.statusPending}</span>
    if (status === 'approved') return <span className="ard-badge ard-badge-approved">{t.userManagement.statusApproved}</span>
    if (status === 'denied')   return <span className="ard-badge ard-badge-denied">{t.userManagement.statusDenied}</span>
    return <span className="ard-badge">{status}</span>
  }

  const filteredRequests = requests.filter(req => {
    const matchesStatus = filterStatus === 'all' || req.approvalStatus === filterStatus
    const matchesSearch = searchTerm === '' ||
      req.email.toLowerCase().includes(searchTerm.toLowerCase()) ||
      req.desiredRoles.some(role => role.toLowerCase().includes(searchTerm.toLowerCase()))
    return matchesStatus && matchesSearch
  })

  return (
    <div className="ard-page">
      <div className="orb orb-1" />
      <div className="orb orb-2" />

      <div className="ard-container">
        <div className="ard-header">
          <div>
            <h2 className="ard-title">{t.userManagement.title}</h2>
            <p className="ard-subtitle">{t.userManagement.subtitle}</p>
          </div>
        </div>

        <div className="ard-controls">
          <div className="ard-search">
            <input
              type="text"
              placeholder={t.userManagement.searchPlaceholder}
              value={searchTerm}
              onChange={e => setSearchTerm(e.target.value)}
              className="ard-search-input"
            />
            <svg className="ard-search-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <circle cx="11" cy="11" r="8"/>
              <path d="m21 21-4.35-4.35"/>
            </svg>
          </div>

          <div className="ard-filter">
            <select
              value={filterStatus}
              onChange={e => setFilterStatus(e.target.value)}
              className="ard-filter-select"
            >
              <option value="all">{t.userManagement.filterAll}</option>
              <option value="pending">{t.userManagement.statusPending}</option>
              <option value="approved">{t.userManagement.statusApproved}</option>
              <option value="denied">{t.userManagement.statusDenied}</option>
            </select>
          </div>
        </div>

        {loading && <p className="ard-info">{t.userManagement.loading}</p>}
        {error && <p className="ard-error">{error}</p>}

        {!loading && !error && filteredRequests.length === 0 && (
          <p className="ard-info">{t.userManagement.noRequests}</p>
        )}

        {filteredRequests.length > 0 && (
          <div className="ard-table-wrap">
            <table className="ard-table">
              <thead>
                <tr>
                  <th>{t.userManagement.email}</th>
                  <th>{t.userManagement.roles}</th>
                  <th>{t.userManagement.status}</th>
                  <th>{t.userManagement.submitted}</th>
                  <th></th>
                </tr>
              </thead>
              <tbody>
                {filteredRequests.map(req => (
                  <tr key={req.id} className={`ard-row ard-row-${req.approvalStatus}`}>
                    <td>{req.email}</td>
                    <td>{req.desiredRoles?.join(', ')}</td>
                    <td>{statusLabel(req.approvalStatus)}</td>
                    <td>{new Date(req.createdAt).toLocaleString()}</td>
                    <td className="ard-actions">
                      {req.approvalStatus === 'pending' && (
                        <>
                          <button
                            className="ard-btn ard-btn-approve"
                            disabled={acting !== null}
                            onClick={() => act(req.id, 'approve')}
                          >
                            {t.userManagement.approve}
                          </button>
                          <button
                            className="ard-btn ard-btn-reject"
                            disabled={acting !== null}
                            onClick={() => act(req.id, 'reject')}
                          >
                            {t.userManagement.deny}
                          </button>
                        </>
                      )}
                      {/* Cancel is only valid for pending workflows */}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}

        <button className="page-back-btn" onClick={onBack}>
          ← {t.userManagement.back}
        </button>
      </div>
    </div>
  )
}