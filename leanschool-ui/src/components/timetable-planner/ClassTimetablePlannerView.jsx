import { useState, useEffect, useCallback } from 'react'
import { useTranslation } from '../../i18n/useTranslation'
import { useTimetableApi } from '../../hooks/useTimetableApi'
import TimetableGrid from './TimetableGrid'

export default function ClassTimetablePlannerView({ planId, classId, className: cls, conflicts = [] }) {
  const { t } = useTranslation()
  const api = useTimetableApi()
  const tt = t.timetable || {}

  const [entries, setEntries] = useState([])
  const [timeSlots, setTimeSlots] = useState([])
  const [teachers, setTeachers] = useState([])
  const [rooms, setRooms] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [selected, setSelected] = useState(null)

  const conflictEntryIds = new Set(
    conflicts.flatMap(c => c.entryIds || [])
  )

  const loadEntries = useCallback(async () => {
    try {
      const e = await api.listEntries(planId, { classId })
      setEntries(e || [])
    } catch (e) { setError(e.message) }
  }, [planId, classId]) // eslint-disable-line react-hooks/exhaustive-deps

  useEffect(() => {
    Promise.all([
      api.listEntries(planId, { classId }),
      api.listTimeSlots(planId),
      api.listTeachers(planId),
      api.listRooms(planId),
    ])
      .then(([e, ts, t, r]) => {
        setEntries(e || []); setTimeSlots(ts || [])
        setTeachers(t || []); setRooms(r || [])
      })
      .catch(e => setError(e.message))
      .finally(() => setLoading(false))
  }, [planId, classId]) // eslint-disable-line react-hooks/exhaustive-deps

  async function handleSwap(entryId, targetEntryId) {
    try {
      await api.swapEntries(planId, entryId, targetEntryId)
      await loadEntries()
    } catch (e) { setError(e.message) }
  }

  async function handleReassign(entryId, teacherId) {
    try {
      await api.reassignTeacher(planId, entryId, teacherId)
      await loadEntries()
      setSelected(null)
    } catch (e) { setError(e.message) }
  }

  if (loading) return <div className="dv-loading"><div className="spinner" /></div>

  return (
    <div style={{ display: 'flex', flex: 1, gap: 16, padding: 16, minHeight: 0 }}>
      {/* Grid */}
      <div style={{ flex: 1, display: 'flex', flexDirection: 'column', minWidth: 0 }}>
        {cls && <h3 style={{ color: 'var(--text-1)', marginBottom: 8 }}>{cls}</h3>}
        {error && <div style={{ color: 'var(--error)', marginBottom: 8, fontSize: 13 }}>{error}</div>}
        <TimetableGrid
          entries={entries}
          timeSlots={timeSlots}
          readOnly={false}
          highlightConflicts={conflictEntryIds}
          onEntryClick={setSelected}
          onSwap={handleSwap}
        />
      </div>

      {/* Side panel */}
      {selected && (
        <div style={{
          width: 260, background: 'var(--bg-card)', borderRadius: 8,
          border: '1px solid var(--border-1)', padding: 16,
          display: 'flex', flexDirection: 'column', gap: 12, flexShrink: 0,
        }}>
          <h4 style={{ margin: 0, color: 'var(--text-1)' }}>{selected.subjectName}</h4>
          <div style={{ fontSize: 13, color: 'var(--text-2)' }}>
            {tt.teacher || 'Teacher'}: {selected.teacherName || '—'}
          </div>
          <div style={{ fontSize: 13, color: 'var(--text-2)' }}>
            {tt.room || 'Room'}: {selected.roomName || '—'}
          </div>

          <label style={{ fontSize: 12, color: 'var(--text-2)' }}>
            {tt.reassignTeacher || 'Reassign Teacher'}
            <select
              className="field-input"
              value={selected.teacherId || ''}
              onChange={e => handleReassign(selected.id, e.target.value)}
              style={{ width: '100%', marginTop: 4 }}
            >
              <option value="">—</option>
              {teachers.map(t => (
                <option key={t.id} value={t.id}>{t.prename} {t.name}</option>
              ))}
            </select>
          </label>

          <button className="ghost-button" onClick={() => setSelected(null)} style={{ marginTop: 'auto' }}>
            Close
          </button>
        </div>
      )}
    </div>
  )
}
