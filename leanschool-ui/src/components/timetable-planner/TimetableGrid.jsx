import { useState } from 'react'
import { DndContext, DragOverlay, PointerSensor, useSensor, useSensors } from '@dnd-kit/core'
import { useDroppable } from '@dnd-kit/core'
import { useTranslation } from '../../i18n/useTranslation'
import EntryCard, { EntryCardOverlay } from './EntryCard'
import './TimetableGrid.css'

const DAYS = [1, 2, 3, 4, 5]
const DAY_KEYS = { 1: 'monday', 2: 'tuesday', 3: 'wednesday', 4: 'thursday', 5: 'friday' }

function DroppableCell({ id, children }) {
  const { setNodeRef, isOver } = useDroppable({ id })
  return (
    <div ref={setNodeRef} className={`tg-cell${isOver ? ' tg-cell-over' : ''}`}>
      {children}
    </div>
  )
}

export default function TimetableGrid({
  entries = [],
  timeSlots = [],
  readOnly = true,
  highlightConflicts = new Set(),
  onEntryClick,
  onSwap,
}) {
  const { t } = useTranslation()
  const [activeEntry, setActiveEntry] = useState(null)

  const sensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: 5 } })
  )

  // Group time slots by day and sort by period
  const periods = [...new Set(timeSlots.map(ts => ts.period))].sort((a, b) => a - b)
  const slotMap = {}
  for (const ts of timeSlots) {
    slotMap[`${ts.dayOfWeek}-${ts.period}`] = ts
  }

  // Build entry lookup: timeSlotId → entry
  const entryBySlot = {}
  for (const e of entries) {
    entryBySlot[e.timeSlotId] = e
  }

  function handleDragStart(event) {
    const entry = event.active.data.current?.entry
    if (entry) setActiveEntry(entry)
  }

  function handleDragEnd(event) {
    setActiveEntry(null)
    const { active, over } = event
    if (!over || !active) return

    const draggedEntry = active.data.current?.entry
    if (!draggedEntry) return

    // over.id is either a slot cell ID "cell-{tsId}" or another entry ID
    const overId = String(over.id)
    if (overId.startsWith('cell-')) {
      const targetSlotId = overId.replace('cell-', '')
      const targetEntry = Object.values(entryBySlot).find(e => e.timeSlotId === targetSlotId)
      if (targetEntry && targetEntry.id !== draggedEntry.id) {
        onSwap?.(draggedEntry.id, targetEntry.id)
      }
    }
  }

  const grid = (
    <div className="tg-grid" style={{ gridTemplateColumns: `80px repeat(${DAYS.length}, 1fr)`, gridTemplateRows: `auto repeat(${periods.length}, 1fr)` }}>
      {/* Header row */}
      <div className="tg-header tg-corner" />
      {DAYS.map(d => (
        <div key={d} className="tg-header tg-day-header">
          {t.timetable?.[DAY_KEYS[d]] || DAY_KEYS[d]}
        </div>
      ))}

      {/* Period rows */}
      {periods.map(p => {
        // Find a representative slot for this period to get the time
        const sampleSlot = timeSlots.find(ts => ts.period === p)
        return [
          <div key={`period-${p}`} className="tg-period-label">
            <span className="tg-period-num">{p}</span>
            {sampleSlot && (
              <span className="tg-period-time">{sampleSlot.startTime}–{sampleSlot.endTime}</span>
            )}
          </div>,
          ...DAYS.map(d => {
            const ts = slotMap[`${d}-${p}`]
            if (!ts) return <div key={`empty-${d}-${p}`} className="tg-cell tg-cell-disabled" />
            const entry = entryBySlot[ts.id]
            const cellId = `cell-${ts.id}`
            return (
              <DroppableCell key={cellId} id={cellId}>
                {entry && (
                  <EntryCard
                    entry={entry}
                    readOnly={readOnly}
                    isConflict={highlightConflicts.has(entry.id)}
                    onClick={onEntryClick}
                  />
                )}
              </DroppableCell>
            )
          }),
        ]
      })}
    </div>
  )

  if (readOnly) return grid

  return (
    <DndContext sensors={sensors} onDragStart={handleDragStart} onDragEnd={handleDragEnd}>
      {grid}
      <DragOverlay>
        {activeEntry ? <EntryCardOverlay entry={activeEntry} /> : null}
      </DragOverlay>
    </DndContext>
  )
}
