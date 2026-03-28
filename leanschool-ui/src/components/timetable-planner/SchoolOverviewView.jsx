import { useState, useEffect } from 'react'
import { useTranslation } from '../../i18n/useTranslation'
import { useTimetableApi } from '../../hooks/useTimetableApi'
import './SchoolOverviewView.css'

export default function SchoolOverviewView({ plan, onSelectClass, onFinalized }) {
  const { t } = useTranslation()
  const api = useTimetableApi()
  const tt = t.timetable || {}

  const [classes, setClasses] = useState([])
  const [conflicts, setConflicts] = useState([])
  const [loading, setLoading] = useState(true)
  const [finalizing, setFinalizing] = useState(false)
  const [error, setError] = useState(null)
  const [confirmFinalize, setConfirmFinalize] = useState(false)

  useEffect(() => {
    Promise.all([
      api.listClasses(plan.id),
      api.listConflicts(plan.id),
    ])
      .then(([cls, conf]) => {
        setClasses(cls || [])
        setConflicts(conf || [])
      })
      .catch(e => setError(e.message))
      .finally(() => setLoading(false))
  }, [plan.id]) // eslint-disable-line react-hooks/exhaustive-deps

  // Count conflicts per class
  const conflictsByClass = {}
  for (const c of conflicts) {
    if (c.schoolClassId) {
      conflictsByClass[c.schoolClassId] = (conflictsByClass[c.schoolClassId] || 0) + 1
    }
  }

  async function handleFinalize() {
    setFinalizing(true); setError(null)
    try {
      const result = await api.finalize(plan.id)
      onFinalized?.(result)
    } catch (e) { setError(e.message) }
    finally { setFinalizing(false); setConfirmFinalize(false) }
  }

  if (loading) return <div className="dv-loading"><div className="spinner" /></div>

  const canFinalize = plan.status === 'accepted'

  return (
    <div style={{ padding: 24 }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
        <h2 style={{ margin: 0, color: 'var(--text-1)' }}>{tt.overview || 'School Overview'}</h2>
        <div style={{ display: 'flex', gap: 8, alignItems: 'center' }}>
          <span style={{ fontSize: 13, color: conflicts.length === 0 ? 'var(--ok)' : 'var(--error)' }}>
            {conflicts.length} {tt.conflicts || 'conflicts'}
          </span>
          {canFinalize && !confirmFinalize && (
            <button className="cta-button" onClick={() => setConfirmFinalize(true)}>
              {tt.acceptTimetable || 'Accept & Finalize'}
            </button>
          )}
          {confirmFinalize && (
            <div style={{ display: 'flex', gap: 8, alignItems: 'center' }}>
              <span style={{ fontSize: 13, color: 'var(--warn)' }}>Are you sure?</span>
              <button className="cta-button" onClick={handleFinalize} disabled={finalizing}>
                {finalizing ? '...' : 'Confirm'}
              </button>
              <button className="ghost-button" onClick={() => setConfirmFinalize(false)}>Cancel</button>
            </div>
          )}
        </div>
      </div>

      {error && <div style={{ color: 'var(--error)', marginBottom: 12 }}>{error}</div>}

      <div className="sov-grid">
        {classes.map(cls => {
          const count = conflictsByClass[cls.id] || 0
          return (
            <div
              key={cls.id}
              className="sov-card"
              onClick={() => onSelectClass?.(cls)}
            >
              <div className="sov-card-name">{cls.name}</div>
              {cls.shortcut && <div className="sov-card-shortcut">{cls.shortcut}</div>}
              {count > 0 ? (
                <span className="sov-badge sov-badge-error">{count}</span>
              ) : (
                <span className="sov-badge sov-badge-ok">0</span>
              )}
            </div>
          )
        })}
      </div>
    </div>
  )
}
