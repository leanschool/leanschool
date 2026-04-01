import { useState, useEffect } from 'react';
import { useAuth } from '../../auth/useAuth';
import { useTranslation } from '../../i18n/useTranslation';
import './domain-components.css';
import { config } from '../../config';

const API = config.leanschoolUrl;

export default function LessonForm({ id, persist = false, onSave, onCancel }) {
  const { t } = useTranslation();
  const { authFetch } = useAuth();
  const [teacherId, setTeacherId] = useState('');
  const [schoolClassId, setSchoolClassId] = useState('');
  const [subjectId, setSubjectId] = useState('');
  const [roomId, setRoomId] = useState('');
  const [version, setVersion] = useState(0);
  const [teachers, setTeachers] = useState([]);
  const [schoolClasses, setSchoolClasses] = useState([]);
  const [subjects, setSubjects] = useState([]);
  const [rooms, setRooms] = useState([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState(null);
  const [conflict, setConflict] = useState(false);

  useEffect(() => {
    const fetches = [
      authFetch(`${API}/teachers`).then(res => res.ok ? res.json() : []).catch(() => []),
      authFetch(`${API}/school-classes`).then(res => res.ok ? res.json() : []).catch(() => []),
      authFetch(`${API}/subjects`).then(res => res.ok ? res.json() : []).catch(() => []),
      authFetch(`${API}/rooms`).then(res => res.ok ? res.json() : []).catch(() => []),
    ];

    if (id) {
      fetches.push(
        authFetch(`${API}/lessons/${id}`)
          .then(res => {
            if (!res.ok) throw new Error(`Failed to load lesson (${res.status})`);
            return res.json();
          })
      );
    }

    Promise.all(fetches)
      .then(([teacherList, classList, subjectList, roomList, lesson]) => {
        setTeachers(teacherList || []);
        setSchoolClasses(classList || []);
        setSubjects(subjectList || []);
        setRooms(roomList || []);
        if (lesson) {
          setTeacherId(lesson.teacher?.id ?? '');
          setSchoolClassId(lesson.schoolClass?.id ?? '');
          setSubjectId(lesson.subject?.id ?? '');
          setRoomId(lesson.room?.id ?? '');
          setVersion(lesson.version);
        }
      })
      .catch(err => setError(err.message))
      .finally(() => setLoading(false));
  }, [id]);

  async function handleSubmit(e) {
    e.preventDefault();
    setConflict(false);
    setError(null);

    const body = {
      ...(teacherId ? { teacher: { id: teacherId } } : {}),
      ...(schoolClassId ? { schoolClass: { id: schoolClassId } } : {}),
      ...(subjectId ? { subject: { id: subjectId } } : {}),
      ...(roomId ? { room: { id: roomId } } : {}),
      ...(id ? { version } : {}),
    };

    if (!persist) {
      onSave?.(body);
      return;
    }

    setSaving(true);
    const url = id ? `${API}/lessons/${id}` : `${API}/lessons`;
    const method = id ? 'PUT' : 'POST';

    try {
      const res = await authFetch(url, {
        method,
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body),
      });

      if (res.status === 409) { setConflict(true); return; }
      if (!res.ok) throw new Error(res.status);
      const saved = await res.json();
      onSave?.(saved);
    } catch {
      setError(t.domain.common.saveError);
    } finally {
      setSaving(false);
    }
  }

  if (loading) {
    return (
      <div className="dv-loading">
        <div className="spinner" />
      </div>
    );
  }

  return (
    <form className="df-form" onSubmit={handleSubmit}>
      <div className="df-grid">
        <div className="df-field-group">
          <div className="df-label">{t.domain.lesson.teacher}</div>
          <select
            className="field-input field-select"
            value={teacherId}
            onChange={e => setTeacherId(e.target.value)}
          >
            <option value="">{t.domain.common.none}</option>
            {teachers.map(t => (
              <option key={t.id} value={t.id}>
                {t.prename} {t.name}
              </option>
            ))}
          </select>
        </div>
        <div className="df-field-group">
          <div className="df-label">{t.domain.lesson.schoolClass}</div>
          <select
            className="field-input field-select"
            value={schoolClassId}
            onChange={e => setSchoolClassId(e.target.value)}
          >
            <option value="">{t.domain.common.none}</option>
            {schoolClasses.map(sc => (
              <option key={sc.id} value={sc.id}>
                {sc.name ?? sc.id}
              </option>
            ))}
          </select>
        </div>
        <div className="df-field-group">
          <div className="df-label">{t.domain.lesson.subject}</div>
          <select
            className="field-input field-select"
            value={subjectId}
            onChange={e => setSubjectId(e.target.value)}
          >
            <option value="">{t.domain.common.none}</option>
            {subjects.map(s => (
              <option key={s.id} value={s.id}>
                {s.name ?? s.id}
              </option>
            ))}
          </select>
        </div>
        <div className="df-field-group">
          <div className="df-label">{t.domain.lesson.room}</div>
          <select
            className="field-input field-select"
            value={roomId}
            onChange={e => setRoomId(e.target.value)}
          >
            <option value="">{t.domain.common.none}</option>
            {rooms.map(r => (
              <option key={r.id} value={r.id}>
                {r.name}{r.roomType ? ` (${t.domain.roomTypes[r.roomType] || r.roomType})` : ''}
              </option>
            ))}
          </select>
        </div>
      </div>

      {conflict && (
        <div className="df-conflict-error">
          {t.domain.common.conflict}
        </div>
      )}
      {error && <div className="dv-error">{error}</div>}

      <div className="df-actions">
        <button type="submit" className="cta-button am-btn" disabled={saving}>
          {saving ? t.domain.common.saving : id ? t.domain.common.update : t.domain.common.create}
        </button>
        {onCancel && (
          <button type="button" className="ghost-button am-btn" onClick={onCancel}>
            {t.domain.common.cancel}
          </button>
        )}
      </div>
    </form>
  );
}
