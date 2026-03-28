import { useState, useEffect } from 'react'
import { useTranslation } from '../../i18n/useTranslation'
import { useTimetableApi } from '../../hooks/useTimetableApi'
import './PlanSetup.css'

export default function PlanSetup({ plan, onGenerated, onPlanUpdated }) {
  const { t } = useTranslation()
  const api = useTimetableApi()
  const tt = t.timetable || {}

  const [loading, setLoading] = useState(false)
  const [message, setMessage] = useState(null)
  const [error, setError] = useState(null)

  // Snapshot state
  const [snapshotDone, setSnapshotDone] = useState(false)
  const [snapshotSummary, setSnapshotSummary] = useState(null)

  // Time slot defaults
  const [slotConfig, setSlotConfig] = useState({
    morningPeriods: 4, afternoonPeriods: 3,
    startTime: '08:00', lessonDurationMin: 45,
    breakDurationMin: 5, lunchBreakMin: 60,
  })
  const [slotsGenerated, setSlotsGenerated] = useState(false)
  const [timeSlots, setTimeSlots] = useState([])

  // Requirements & constraints
  const [requirements, setRequirements] = useState([])
  const [constraints, setConstraints] = useState([])
  const [classes, setClasses] = useState([])
  const [subjects, setSubjects] = useState([])

  // Requirement form
  const [reqForm, setReqForm] = useState({ schoolClassId: '', subjectId: '', lessonsPerWeek: 3, maxDoubleLessons: 1, preferMorning: false })
  // Constraint form
  const [conForm, setConForm] = useState({ schoolClassId: '', maxEarlyStarts: 3, morningPeriods: 4, afternoonPeriods: 3, freeAfternoons: 1, freeAfternoonDays: '', hasTimetable: true })

  useEffect(() => {
    loadData()
  }, [plan.id]) // eslint-disable-line react-hooks/exhaustive-deps

  async function loadData() {
    try {
      const [ts, reqs, cons, cls, subs] = await Promise.all([
        api.listTimeSlots(plan.id),
        api.listRequirements(plan.id),
        api.listConstraints(plan.id),
        api.listClasses(plan.id).catch(() => []),
        api.listSubjects(plan.id).catch(() => []),
      ])
      setTimeSlots(ts || [])
      setSlotsGenerated((ts || []).length > 0)
      setRequirements(reqs || [])
      setConstraints(cons || [])
      setClasses(cls || [])
      setSubjects(subs || [])
      if ((cls || []).length > 0) setSnapshotDone(true)
    } catch (e) {
      setError(e.message)
    }
  }

  async function handleSnapshot() {
    setLoading(true); setError(null)
    try {
      const summary = await api.takeSnapshot(plan.id)
      setSnapshotSummary(summary)
      setSnapshotDone(true)
      await loadData()
    } catch (e) { setError(e.message) }
    finally { setLoading(false) }
  }

  async function handleGenerateSlots() {
    setLoading(true); setError(null)
    try {
      await api.generateDefaultSlots(plan.id, slotConfig)
      setSlotsGenerated(true)
      const ts = await api.listTimeSlots(plan.id)
      setTimeSlots(ts || [])
      setMessage(`Generated ${(ts || []).length} time slots`)
    } catch (e) { setError(e.message) }
    finally { setLoading(false) }
  }

  async function handleAddRequirement() {
    if (!reqForm.schoolClassId || !reqForm.subjectId) return
    setLoading(true); setError(null)
    try {
      const sub = subjects.find(s => s.id === reqForm.subjectId)
      await api.createRequirement(plan.id, { ...reqForm, subjectName: sub?.name || '' })
      const reqs = await api.listRequirements(plan.id)
      setRequirements(reqs || [])
      setReqForm({ schoolClassId: '', subjectId: '', lessonsPerWeek: 3, maxDoubleLessons: 1, preferMorning: false })
    } catch (e) { setError(e.message) }
    finally { setLoading(false) }
  }

  async function handleDeleteRequirement(id) {
    try {
      await api.deleteRequirement(plan.id, id)
      setRequirements(prev => prev.filter(r => r.id !== id))
    } catch (e) { setError(e.message) }
  }

  async function handleAddConstraint() {
    if (!conForm.schoolClassId) return
    setLoading(true); setError(null)
    try {
      const cls = classes.find(c => c.id === conForm.schoolClassId)
      await api.createConstraint(plan.id, { ...conForm, schoolClassName: cls?.name || '' })
      const cons = await api.listConstraints(plan.id)
      setConstraints(cons || [])
      setConForm({ schoolClassId: '', maxEarlyStarts: 3, morningPeriods: 4, afternoonPeriods: 3, freeAfternoons: 1, freeAfternoonDays: '', hasTimetable: true })
    } catch (e) { setError(e.message) }
    finally { setLoading(false) }
  }

  async function handleDeleteConstraint(id) {
    try {
      await api.deleteConstraint(plan.id, id)
      setConstraints(prev => prev.filter(c => c.id !== id))
    } catch (e) { setError(e.message) }
  }

  async function handleGenerate() {
    setLoading(true); setError(null)
    try {
      const result = await api.generate(plan.id)
      setMessage(`Generated ${result.entries} entries with ${result.conflicts} conflicts`)
      const updated = await api.getPlan(plan.id)
      onPlanUpdated?.(updated)
      onGenerated?.(result)
    } catch (e) { setError(e.message) }
    finally { setLoading(false) }
  }

  return (
    <div className="ps-container">
      <h2 className="ps-title">{tt.setup || 'Setup'}: {plan.name}</h2>
      {error && <div className="ps-error">{error}</div>}
      {message && <div className="ps-message">{message}</div>}

      {/* Step 1: Snapshot */}
      <section className="ps-section">
        <h3 className="ps-section-heading">1. {tt.snapshot || 'Take Snapshot'}</h3>
        <p className="ps-desc">{tt.snapshotDesc || 'Capture current teachers, subjects, classes, and rooms'}</p>
        <button className="cta-button" onClick={handleSnapshot} disabled={loading}>
          {tt.snapshot || 'Take Snapshot'}
        </button>
        {snapshotSummary && (
          <span className="ps-snapshot-summary">
            {snapshotSummary.teachers} teachers, {snapshotSummary.subjects} subjects, {snapshotSummary.classes} classes, {snapshotSummary.rooms} rooms
          </span>
        )}
      </section>

      {/* Step 2: Time Slots */}
      {snapshotDone && (
        <section className="ps-section">
          <h3 className="ps-section-heading">2. {tt.timeSlots || 'Time Slots'}</h3>
          {!slotsGenerated ? (
            <div className="ps-slot-row">
              <label className="ps-slot-label">
                {tt.morning || 'Morning'} Periods
                <input type="number" className="field-input ps-input-narrow" value={slotConfig.morningPeriods} min={1} max={8}
                  onChange={e => setSlotConfig(p => ({ ...p, morningPeriods: +e.target.value }))} />
              </label>
              <label className="ps-slot-label">
                {tt.afternoon || 'Afternoon'} Periods
                <input type="number" className="field-input ps-input-narrow" value={slotConfig.afternoonPeriods} min={0} max={6}
                  onChange={e => setSlotConfig(p => ({ ...p, afternoonPeriods: +e.target.value }))} />
              </label>
              <label className="ps-slot-label">
                Start Time
                <input type="time" className="field-input" value={slotConfig.startTime}
                  onChange={e => setSlotConfig(p => ({ ...p, startTime: e.target.value }))} />
              </label>
              <label className="ps-slot-label">
                Lesson (min)
                <input type="number" className="field-input ps-input-narrow" value={slotConfig.lessonDurationMin} min={15} max={120}
                  onChange={e => setSlotConfig(p => ({ ...p, lessonDurationMin: +e.target.value }))} />
              </label>
              <label className="ps-slot-label">
                Break (min)
                <input type="number" className="field-input ps-input-narrow" value={slotConfig.breakDurationMin} min={0} max={30}
                  onChange={e => setSlotConfig(p => ({ ...p, breakDurationMin: +e.target.value }))} />
              </label>
              <label className="ps-slot-label">
                Lunch (min)
                <input type="number" className="field-input ps-input-narrow" value={slotConfig.lunchBreakMin} min={0} max={120}
                  onChange={e => setSlotConfig(p => ({ ...p, lunchBreakMin: +e.target.value }))} />
              </label>
              <button className="cta-button" onClick={handleGenerateSlots} disabled={loading}>
                {tt.generateDefault || 'Generate Default Grid'}
              </button>
            </div>
          ) : (
            <p className="ps-slots-ok">{timeSlots.length} time slots configured</p>
          )}
        </section>
      )}

      {/* Step 3: Requirements */}
      {slotsGenerated && (
        <section className="ps-section">
          <h3 className="ps-section-heading">3. {tt.requirements || 'Requirements'}</h3>
          {requirements.length > 0 && (
            <table className="ps-table">
              <thead>
                <tr className="ps-thead-row">
                  <th className="ps-th-left">Class</th>
                  <th className="ps-th-left">{tt.subject || 'Subject'}</th>
                  <th className="ps-th">{tt.lessonsPerWeek || 'L/W'}</th>
                  <th className="ps-th">Max Dbl</th>
                  <th className="ps-th"></th>
                </tr>
              </thead>
              <tbody>
                {requirements.map(r => (
                  <tr key={r.id} className="ps-tbody-row">
                    <td className="ps-td">{classes.find(c => c.id === r.schoolClassId)?.name || r.schoolClassId}</td>
                    <td className="ps-td">{r.subjectName || r.subjectId}</td>
                    <td className="ps-td-center">{r.lessonsPerWeek}</td>
                    <td className="ps-td-center">{r.maxDoubleLessons}</td>
                    <td className="ps-td"><button className="ghost-button ps-delete-btn" onClick={() => handleDeleteRequirement(r.id)}>x</button></td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
          <div className="ps-form-row">
            <select className="field-input" value={reqForm.schoolClassId} onChange={e => setReqForm(p => ({ ...p, schoolClassId: e.target.value }))}>
              <option value="">Class...</option>
              {classes.map(c => <option key={c.id} value={c.id}>{c.name}</option>)}
            </select>
            <select className="field-input" value={reqForm.subjectId} onChange={e => setReqForm(p => ({ ...p, subjectId: e.target.value }))}>
              <option value="">Subject...</option>
              {subjects.map(s => <option key={s.id} value={s.id}>{s.name}</option>)}
            </select>
            <input type="number" className="field-input ps-input-small" value={reqForm.lessonsPerWeek} min={1} max={20}
              onChange={e => setReqForm(p => ({ ...p, lessonsPerWeek: +e.target.value }))} title={tt.lessonsPerWeek} />
            <input type="number" className="field-input ps-input-small" value={reqForm.maxDoubleLessons} min={0} max={10}
              onChange={e => setReqForm(p => ({ ...p, maxDoubleLessons: +e.target.value }))} title={tt.maxDoubleLessons} />
            <button className="cta-button ps-add-btn" onClick={handleAddRequirement} disabled={loading}>+</button>
          </div>
        </section>
      )}

      {/* Step 4: Constraints */}
      {slotsGenerated && (
        <section className="ps-section">
          <h3 className="ps-section-heading">4. {tt.constraints || 'Constraints'}</h3>
          {constraints.length > 0 && (
            <table className="ps-table">
              <thead>
                <tr className="ps-thead-row">
                  <th className="ps-th-left">Class</th>
                  <th className="ps-th">Early</th>
                  <th className="ps-th">Free PM</th>
                  <th className="ps-th">Timetable</th>
                  <th className="ps-th"></th>
                </tr>
              </thead>
              <tbody>
                {constraints.map(c => (
                  <tr key={c.id} className="ps-tbody-row">
                    <td className="ps-td">{c.schoolClassName || c.schoolClassId}</td>
                    <td className="ps-td-center">{c.maxEarlyStarts}</td>
                    <td className="ps-td-center">{c.freeAfternoons}</td>
                    <td className="ps-td-center">{c.hasTimetable ? 'Yes' : 'No'}</td>
                    <td className="ps-td"><button className="ghost-button ps-delete-btn" onClick={() => handleDeleteConstraint(c.id)}>x</button></td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
          <div className="ps-form-row">
            <select className="field-input" value={conForm.schoolClassId} onChange={e => setConForm(p => ({ ...p, schoolClassId: e.target.value }))}>
              <option value="">Class...</option>
              {classes.map(c => <option key={c.id} value={c.id}>{c.name}</option>)}
            </select>
            <input type="number" className="field-input ps-input-small" value={conForm.maxEarlyStarts} min={0} max={5}
              onChange={e => setConForm(p => ({ ...p, maxEarlyStarts: +e.target.value }))} title={tt.maxEarlyStarts} />
            <input type="number" className="field-input ps-input-small" value={conForm.freeAfternoons} min={0} max={5}
              onChange={e => setConForm(p => ({ ...p, freeAfternoons: +e.target.value }))} title={tt.freeAfternoons} />
            <label className="ps-checkbox-label">
              <input type="checkbox" checked={conForm.hasTimetable}
                onChange={e => setConForm(p => ({ ...p, hasTimetable: e.target.checked }))} />
              {tt.hasTimetable || 'Has Timetable'}
            </label>
            <button className="cta-button ps-add-btn" onClick={handleAddConstraint} disabled={loading}>+</button>
          </div>
        </section>
      )}

      {/* Step 5: Generate */}
      {slotsGenerated && requirements.length > 0 && constraints.length > 0 && (
        <section>
          <h3 className="ps-section-heading">5. {tt.generate || 'Generate Timetable'}</h3>
          <button className="cta-button" onClick={handleGenerate} disabled={loading}>
            {tt.generate || 'Generate Timetable'}
          </button>
        </section>
      )}
    </div>
  )
}
