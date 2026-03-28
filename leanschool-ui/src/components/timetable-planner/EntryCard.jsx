import { useDraggable } from '@dnd-kit/core'
import { CSS } from '@dnd-kit/utilities'
import './EntryCard.css'

function subjectColor(name) {
  let h = 0
  for (let i = 0; i < name.length; i++) h = (h * 31 + name.charCodeAt(i)) & 0xffffff
  return `hsl(${h % 360}, 55%, 60%)`
}

export default function EntryCard({ entry, readOnly, isConflict, onClick }) {
  const { attributes, listeners, setNodeRef, transform, isDragging } = useDraggable({
    id: entry.id,
    disabled: readOnly,
    data: { entry },
  })

  const style = {
    transform: CSS.Translate.toString(transform),
    opacity: isDragging ? 0.4 : 1,
  }

  return (
    <div
      ref={setNodeRef}
      className={`ec-card${isConflict ? ' ec-conflict' : ''}${isDragging ? ' ec-dragging' : ''}`}
      style={style}
      onClick={() => onClick?.(entry)}
      {...(readOnly ? {} : { ...listeners, ...attributes })}
    >
      <div className="ec-color-strip" style={{ backgroundColor: subjectColor(entry.subjectName || '') }} />
      <div className="ec-content">
        <div className="ec-subject">{entry.subjectName}</div>
        {entry.teacherName && <div className="ec-teacher">{entry.teacherName}</div>}
        {entry.roomName && <div className="ec-room">{entry.roomName}</div>}
      </div>
    </div>
  )
}

export function EntryCardOverlay({ entry }) {
  return (
    <div className="ec-card ec-overlay">
      <div className="ec-color-strip" style={{ backgroundColor: subjectColor(entry.subjectName || '') }} />
      <div className="ec-content">
        <div className="ec-subject">{entry.subjectName}</div>
        {entry.teacherName && <div className="ec-teacher">{entry.teacherName}</div>}
        {entry.roomName && <div className="ec-room">{entry.roomName}</div>}
      </div>
    </div>
  )
}
