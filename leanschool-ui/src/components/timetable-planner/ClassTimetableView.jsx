import { useState, useEffect } from 'react'
import { useTranslation } from '../../i18n/useTranslation'
import { useTimetableApi } from '../../hooks/useTimetableApi'
import TimetableGrid from './TimetableGrid'

export default function ClassTimetableView({ planId, classId, className: cls }) {
  const { t } = useTranslation()
  const api = useTimetableApi()
  const [entries, setEntries] = useState([])
  const [timeSlots, setTimeSlots] = useState([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    Promise.all([
      api.listEntries(planId, { classId }),
      api.listTimeSlots(planId),
    ])
      .then(([e, ts]) => { setEntries(e || []); setTimeSlots(ts || []) })
      .finally(() => setLoading(false))
  }, [planId, classId]) // eslint-disable-line react-hooks/exhaustive-deps

  if (loading) return <div className="dv-loading"><div className="spinner" /></div>

  return (
    <div style={{ padding: 16, flex: 1, display: 'flex', flexDirection: 'column' }}>
      {cls && <h3 style={{ color: 'var(--text-1)', marginBottom: 8 }}>{cls}</h3>}
      <TimetableGrid entries={entries} timeSlots={timeSlots} readOnly />
    </div>
  )
}
