import { useState, useEffect } from 'react'
import { useTranslation } from '../../i18n/useTranslation'
import { useAuth } from '../../auth/useAuth'
import { hasFeature } from '../../auth/permissions'
import { useTimetableApi } from '../../hooks/useTimetableApi'
import PlanList from './PlanList'
import PlanSetup from './PlanSetup'
import ClassTimetableView from './ClassTimetableView'
import ClassTimetablePlannerView from './ClassTimetablePlannerView'
import ConflictView from './ConflictView'
import SchoolOverviewView from './SchoolOverviewView'
import './TimetablePlanner.css'

export default function TimetablePlanner({ onBack }) {
  const { t } = useTranslation()
  const { user } = useAuth()
  const api = useTimetableApi()
  const tt = t.timetable || {}

  const canWrite = hasFeature(user, 'timetablePlanner')

  const [plan, setPlan] = useState(null)
  const [subView, setSubView] = useState('plan-list')
  const [activeClass, setActiveClass] = useState(null)
  const [classes, setClasses] = useState([])
  const [teachers, setTeachers] = useState([])
  const [conflicts, setConflicts] = useState([])
  const [creating, setCreating] = useState(false)
  const [newPlanName, setNewPlanName] = useState('')

  // Load snapshot data when plan is selected
  useEffect(() => {
    if (!plan) return
    Promise.all([
      api.listClasses(plan.id).catch(() => []),
      api.listTeachers(plan.id).catch(() => []),
      api.listConflicts(plan.id).catch(() => []),
    ]).then(([cls, t, c]) => {
      setClasses(cls || [])
      setTeachers(t || [])
      setConflicts(c || [])
    })
  }, [plan?.id, plan?.status]) // eslint-disable-line react-hooks/exhaustive-deps

  function selectPlan(p) {
    setPlan(p)
    if (p.status === 'draft') setSubView('plan-setup')
    else if (p.status === 'resolving') setSubView('plan-conflicts')
    else if (p.status === 'accepted' || p.status === 'finalized') setSubView('plan-overview')
    else setSubView('plan-grid')
  }

  async function handleCreatePlan() {
    if (!newPlanName.trim()) return
    try {
      const p = await api.createPlan({ name: newPlanName.trim(), schoolYearId: '' })
      setCreating(false)
      setNewPlanName('')
      selectPlan(p)
    } catch { /* ignored */ }
  }

  function handleBackToList() {
    setPlan(null)
    setActiveClass(null)
    setSubView('plan-list')
  }

  async function refreshConflicts() {
    if (!plan) return
    try {
      const result = await api.validate(plan.id)
      setConflicts(result.items || [])
      const updated = await api.getPlan(plan.id)
      setPlan(updated)
    } catch { /* ignored */ }
  }

  // Render main content
  let content
  if (subView === 'plan-list') {
    content = (
      <PlanList
        onSelectPlan={selectPlan}
        onCreatePlan={() => setCreating(true)}
      />
    )
  } else if (subView === 'plan-setup' && plan) {
    content = (
      <PlanSetup
        plan={plan}
        onPlanUpdated={p => { setPlan(p); if (p.status !== 'draft') setSubView('plan-conflicts') }}
        onGenerated={() => setSubView('plan-conflicts')}
      />
    )
  } else if (subView === 'plan-grid' && plan && activeClass) {
    const isEditable = plan.status === 'resolving' || plan.status === 'planning'
    content = isEditable ? (
      <ClassTimetablePlannerView
        planId={plan.id}
        classId={activeClass.id}
        className={activeClass.name}
        conflicts={conflicts.filter(c => c.schoolClassId === activeClass.id)}
      />
    ) : (
      <ClassTimetableView planId={plan.id} classId={activeClass.id} className={activeClass.name} />
    )
  } else if (subView === 'plan-conflicts' && plan) {
    content = (
      <ConflictView
        planId={plan.id}
        teachers={teachers}
        onNavigateToClass={(classId) => {
          const cls = classes.find(c => c.id === classId)
          if (cls) { setActiveClass(cls); setSubView('plan-grid') }
        }}
      />
    )
  } else if (subView === 'plan-overview' && plan) {
    content = (
      <SchoolOverviewView
        plan={plan}
        onSelectClass={(cls) => { setActiveClass(cls); setSubView('plan-grid') }}
        onFinalized={() => {
          api.getPlan(plan.id).then(setPlan)
        }}
      />
    )
  } else {
    content = <div style={{ padding: 24, color: 'var(--text-3)' }}>Select a class from the sidebar</div>
  }

  return (
    <div className="tp-container">
      {/* Sidebar */}
      <nav className="tp-sidebar">
        <button className="tp-back-btn" onClick={plan ? handleBackToList : onBack}>
          ← {plan ? (tt.plans || 'Plans') : (tt.backToDashboard || 'Back')}
        </button>

        {plan && (
          <>
            <div className="tp-plan-name">{plan.name}</div>
            <div className="tp-plan-status" data-status={plan.status}>
              {tt[plan.status] || plan.status}
            </div>

            <div className="tp-nav-section">
              {plan.status === 'draft' && (
                <button className={`tp-nav-btn${subView === 'plan-setup' ? ' tp-nav-active' : ''}`}
                  onClick={() => setSubView('plan-setup')}>
                  {tt.setup || 'Setup'}
                </button>
              )}
              <button className={`tp-nav-btn${subView === 'plan-overview' ? ' tp-nav-active' : ''}`}
                onClick={() => setSubView('plan-overview')}>
                {tt.overview || 'Overview'}
              </button>
              <button className={`tp-nav-btn${subView === 'plan-conflicts' ? ' tp-nav-active' : ''}`}
                onClick={() => setSubView('plan-conflicts')}>
                {tt.conflicts || 'Conflicts'} ({conflicts.length})
              </button>
              {plan.status === 'resolving' && (
                <button className="tp-nav-btn tp-validate-btn" onClick={refreshConflicts}>
                  {tt.validate || 'Validate'}
                </button>
              )}
            </div>

            {classes.length > 0 && (
              <div className="tp-nav-section">
                <div className="tp-nav-label">Classes</div>
                {classes.map(cls => (
                  <button
                    key={cls.id}
                    className={`tp-nav-btn${activeClass?.id === cls.id && subView === 'plan-grid' ? ' tp-nav-active' : ''}`}
                    onClick={() => { setActiveClass(cls); setSubView('plan-grid') }}
                  >
                    {cls.name}
                  </button>
                ))}
              </div>
            )}

            {canWrite && plan.status !== 'finalized' && (
              <div className="tp-nav-section" style={{ marginTop: 'auto' }}>
                <button className="tp-nav-btn tp-reset-btn" onClick={async () => {
                  await api.reset(plan.id)
                  const updated = await api.getPlan(plan.id)
                  setPlan(updated)
                  setSubView('plan-setup')
                }}>
                  {tt.reset || 'Reset'}
                </button>
              </div>
            )}
          </>
        )}
      </nav>

      {/* Main content */}
      <main className="tp-main">
        {content}
      </main>

      {/* Create plan dialog */}
      {creating && (
        <div className="tp-overlay" onClick={() => setCreating(false)}>
          <div className="tp-dialog" onClick={e => e.stopPropagation()}>
            <h3 style={{ margin: '0 0 12px', color: 'var(--text-1)' }}>{tt.createPlan || 'Create Plan'}</h3>
            <input
              className="field-input"
              placeholder={tt.planName || 'Plan Name'}
              value={newPlanName}
              onChange={e => setNewPlanName(e.target.value)}
              onKeyDown={e => e.key === 'Enter' && handleCreatePlan()}
              autoFocus
              style={{ width: '100%', marginBottom: 12 }}
            />
            <div style={{ display: 'flex', gap: 8, justifyContent: 'flex-end' }}>
              <button className="ghost-button" onClick={() => setCreating(false)}>Cancel</button>
              <button className="cta-button" onClick={handleCreatePlan}>Create</button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
