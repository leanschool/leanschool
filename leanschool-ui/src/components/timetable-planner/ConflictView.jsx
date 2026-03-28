import { useState, useEffect } from 'react'
import { useTranslation } from '../../i18n/useTranslation'
import { useAuth } from '../../auth/useAuth'
import { useTimetableApi } from '../../hooks/useTimetableApi'
import './ConflictView.css'

const SEVERITY_COLORS = { error: 'var(--error)', warning: 'var(--warn)' }

export default function ConflictView({ planId, teachers = [], onNavigateToClass }) {
  const { t } = useTranslation()
  const { user } = useAuth()
  const api = useTimetableApi()
  const tt = t.timetable || {}

  const [conflicts, setConflicts] = useState([])
  const [loading, setLoading] = useState(true)
  const [filter, setFilter] = useState('all') // 'all' | 'mine'

  // Find teacher snapshot matching current user's sub
  const myTeacher = teachers.find(t => t.sub === user?.sub)

  useEffect(() => {
    setLoading(true)
    const filters = {}
    if (filter === 'mine' && myTeacher) filters.teacherId = myTeacher.id
    api.listConflicts(planId, filters)
      .then(data => setConflicts(data || []))
      .finally(() => setLoading(false))
  }, [planId, filter]) // eslint-disable-line react-hooks/exhaustive-deps

  if (loading) return <div className="dv-loading"><div className="spinner" /></div>

  return (
    <div className="cv-container">
      <div className="cv-header">
        <h2 className="cv-title">{tt.conflicts || 'Conflicts'}</h2>
        <span className="cv-count">({conflicts.length})</span>
        <div className="cv-filter-group">
          <button
            className={`cv-filter-btn${filter === 'all' ? ' cv-active' : ''}`}
            onClick={() => setFilter('all')}
          >{tt.allConflicts || 'All'}</button>
          {myTeacher && (
            <button
              className={`cv-filter-btn${filter === 'mine' ? ' cv-active' : ''}`}
              onClick={() => setFilter('mine')}
            >{tt.myConflicts || 'Mine'}</button>
          )}
        </div>
      </div>

      {conflicts.length === 0 ? (
        <div className="cv-no-conflicts">
          {tt.noConflicts || 'No conflicts found'}
        </div>
      ) : (
        <div className="cv-list">
          {conflicts.map(c => (
            <div key={c.id} className="cv-card">
              <div className="cv-card-header">
                <span className="cv-severity" style={{ background: SEVERITY_COLORS[c.severity] || 'var(--text-3)' }}>
                  {c.severity}
                </span>
                <span className="cv-type">
                  {tt.conflictTypes?.[c.type] || c.type.replace(/_/g, ' ')}
                </span>
              </div>
              <p className="cv-desc">{c.description}</p>
              {c.entryIds?.length > 0 && (
                <div className="cv-entries">
                  {c.entryIds.map(eid => (
                    <button key={eid} className="cv-entry-link" onClick={() => onNavigateToClass?.(c.schoolClassId, eid)}>
                      Entry {eid.slice(0, 8)}...
                    </button>
                  ))}
                </div>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
