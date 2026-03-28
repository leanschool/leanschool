import { useState, useEffect } from 'react'
import { useTranslation } from '../../i18n/useTranslation'
import { useAuth } from '../../auth/useAuth'
import { hasFeature } from '../../auth/permissions'
import { useTimetableApi } from '../../hooks/useTimetableApi'

const STATUS_COLORS = {
  draft: 'var(--text-3)',
  planning: 'var(--accent)',
  resolving: 'var(--warn)',
  accepted: 'var(--ok)',
  finalized: 'var(--ok)',
}

export default function PlanList({ onSelectPlan, onCreatePlan }) {
  const { t } = useTranslation()
  const { user } = useAuth()
  const api = useTimetableApi()
  const [plans, setPlans] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)

  useEffect(() => {
    api.listPlans()
      .then(setPlans)
      .catch(e => setError(e.message))
      .finally(() => setLoading(false))
  }, []) // eslint-disable-line react-hooks/exhaustive-deps

  const tt = t.timetable || {}
  const canWrite = hasFeature(user, 'timetablePlanner')

  if (loading) return <div className="dv-loading"><div className="spinner" /></div>
  if (error) return <div className="dv-error">{error}</div>

  return (
    <div style={{ padding: 24 }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
        <h2 style={{ margin: 0, color: 'var(--text-1)' }}>{tt.plans || 'Plans'}</h2>
        {canWrite && (
          <button className="cta-button" onClick={onCreatePlan}>
            + {tt.createPlan || 'Create Plan'}
          </button>
        )}
      </div>
      {plans.length === 0 ? (
        <p style={{ color: 'var(--text-3)' }}>No plans yet.</p>
      ) : (
        <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
          {plans.map(plan => (
            <div
              key={plan.id}
              onClick={() => onSelectPlan(plan)}
              style={{
                display: 'flex', alignItems: 'center', gap: 12,
                padding: '12px 16px', background: 'var(--bg-card)',
                borderRadius: 8, cursor: 'pointer',
                border: '1px solid var(--border-1)',
                transition: 'box-shadow 0.12s',
              }}
              onMouseEnter={e => e.currentTarget.style.boxShadow = 'var(--shadow-sm)'}
              onMouseLeave={e => e.currentTarget.style.boxShadow = 'none'}
            >
              <div style={{ flex: 1 }}>
                <div style={{ fontWeight: 600, color: 'var(--text-1)' }}>{plan.name}</div>
                <div style={{ fontSize: 12, color: 'var(--text-3)' }}>
                  {plan.createdAt ? new Date(plan.createdAt).toLocaleDateString() : ''}
                </div>
              </div>
              <span style={{
                padding: '2px 10px', borderRadius: 12, fontSize: 11, fontWeight: 600,
                color: '#fff', background: STATUS_COLORS[plan.status] || 'var(--text-3)',
              }}>
                {tt[plan.status] || plan.status}
              </span>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
