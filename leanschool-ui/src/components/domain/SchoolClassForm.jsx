import { useState, useEffect } from 'react';
import { useAuth } from '../../auth/useAuth';
import { useTranslation } from '../../i18n/useTranslation';
import './domain-components.css';
import { config } from '../../config';

const API = config.leanschoolUrl;

export default function SchoolClassForm({ id, persist = false, onSave, onCancel }) {
  const { t } = useTranslation();
  const { authFetch } = useAuth();
  const [fields, setFields] = useState({ name: '', shortcut: '' });
  const [version, setVersion] = useState(0);
  const [loading, setLoading] = useState(!!id);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState(null);
  const [conflict, setConflict] = useState(false);

  const [rooms, setRooms] = useState([]);
  const [selectedRoom, setSelectedRoom] = useState('');
  const [schoolYears, setSchoolYears] = useState([]);
  const [selectedSchoolYear, setSelectedSchoolYear] = useState('');
  const [teachersList, setTeachersList] = useState([]);
  const [teachers, setTeachers] = useState([]);
  const [teacherSelect, setTeacherSelect] = useState('');
  const [studentsList, setStudentsList] = useState([]);
  const [students, setStudents] = useState([]);
  const [studentSelect, setStudentSelect] = useState('');

  useEffect(() => {
    authFetch(`${API}/rooms`)
      .then(res => res.ok ? res.json() : [])
      .then(data => setRooms(Array.isArray(data) ? data : []))
      .catch(() => {});

    authFetch(`${API}/school-years`)
      .then(res => res.ok ? res.json() : [])
      .then(data => setSchoolYears(Array.isArray(data) ? data : []))
      .catch(() => {});

    authFetch(`${API}/teachers`)
      .then(res => res.ok ? res.json() : [])
      .then(data => setTeachersList(Array.isArray(data) ? data : []))
      .catch(() => {});

    authFetch(`${API}/students`)
      .then(res => res.ok ? res.json() : [])
      .then(data => setStudentsList(Array.isArray(data) ? data : []))
      .catch(() => {});
  }, []);

  useEffect(() => {
    if (!id) return;
    setLoading(true);
    authFetch(`${API}/school-classes/${id}`)
      .then(res => {
        if (!res.ok) throw new Error(`Failed to load school class (${res.status})`);
        return res.json();
      })
      .then(data => {
        setFields({
          name: data.name || '',
          shortcut: data.shortcut || '',
        });
        setVersion(data.version);
        setSelectedRoom(data.classroom?.id || '');
        setSelectedSchoolYear(data.schoolYear?.id || '');
        setTeachers(data.teachers || []);
        setStudents(data.students || []);
      })
      .catch(err => setError(err.message))
      .finally(() => setLoading(false));
  }, [id]);

  function handleChange(e) {
    const { name, value } = e.target;
    setFields(prev => ({ ...prev, [name]: value }));
  }

  function addTeacher() {
    if (!teacherSelect) return;
    const teacher = teachersList.find(t => String(t.id) === String(teacherSelect));
    if (!teacher || teachers.find(t => t.id === teacher.id)) return;
    setTeachers(prev => [...prev, teacher]);
    setTeacherSelect('');
  }

  function removeTeacher(tid) {
    setTeachers(prev => prev.filter(t => t.id !== tid));
  }

  function addStudent() {
    if (!studentSelect) return;
    const student = studentsList.find(s => String(s.id) === String(studentSelect));
    if (!student || students.find(s => s.id === student.id)) return;
    setStudents(prev => [...prev, student]);
    setStudentSelect('');
  }

  function removeStudent(sid) {
    setStudents(prev => prev.filter(s => s.id !== sid));
  }

  async function handleSubmit(e) {
    e.preventDefault();
    setConflict(false);
    setError(null);

    const body = {
      name: fields.name,
      ...(fields.shortcut ? { shortcut: fields.shortcut } : {}),
      ...(selectedRoom ? { classroom: { id: selectedRoom } } : {}),
      ...(selectedSchoolYear ? { schoolYear: { id: selectedSchoolYear } } : {}),
      teachers: teachers.map(t => ({ id: t.id })),
      students: students.map(s => ({ id: s.id })),
      ...(id ? { version } : {}),
    };

    if (!persist) {
      onSave?.(body);
      return;
    }

    setSaving(true);
    const url = id ? `${API}/school-classes/${id}` : `${API}/school-classes`;
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
          <div className="df-label">{t.domain.schoolClass.name}</div>
          <input
            className="field-input"
            type="text"
            name="name"
            value={fields.name}
            onChange={handleChange}
            required
          />
        </div>
        <div className="df-field-group">
          <div className="df-label">{t.domain.schoolClass.shortcut}</div>
          <input
            className="field-input"
            type="text"
            name="shortcut"
            value={fields.shortcut}
            onChange={handleChange}
          />
        </div>
        <div className="df-field-group">
          <div className="df-label">{t.domain.schoolClass.classroom}</div>
          <select
            className="field-input field-select"
            value={selectedRoom}
            onChange={e => setSelectedRoom(e.target.value)}
          >
            <option value="">{t.domain.common.none}</option>
            {rooms.map(r => (
              <option key={r.id} value={r.id}>{r.name || r.id}</option>
            ))}
          </select>
        </div>
        <div className="df-field-group">
          <div className="df-label">{t.domain.schoolClass.schoolYear}</div>
          <select
            className="field-input field-select"
            value={selectedSchoolYear}
            onChange={e => setSelectedSchoolYear(e.target.value)}
          >
            <option value="">{t.domain.common.none}</option>
            {schoolYears.map(sy => (
              <option key={sy.id} value={sy.id}>
                {sy.from ? `${sy.from.slice(0, 10)} – ${sy.to.slice(0, 10)}` : sy.id}
              </option>
            ))}
          </select>
        </div>
      </div>

      <div className="df-section">
        <div className="df-section-title">{t.domain.schoolClass.teachers}</div>
        {teachers.map(teacher => (
          <div key={teacher.id} className="df-item-row">
            <span className="field-input" style={{ cursor: 'default' }}>{teacher.prename} {teacher.name}</span>
            <button type="button" className="df-remove-btn" onClick={() => removeTeacher(teacher.id)}>{t.domain.common.remove}</button>
          </div>
        ))}
        <div className="df-item-row">
          <select
            className="field-input field-select"
            value={teacherSelect}
            onChange={e => setTeacherSelect(e.target.value)}
          >
            <option value="">{t.domain.common.select}</option>
            {teachersList
              .filter(t => !teachers.find(at => at.id === t.id))
              .map(t => (
                <option key={t.id} value={t.id}>{t.prename} {t.name}</option>
              ))}
          </select>
          <button type="button" className="df-add-btn" onClick={addTeacher}>{t.domain.common.add}</button>
        </div>
      </div>

      <div className="df-section">
        <div className="df-section-title">{t.domain.schoolClass.students}</div>
        {students.map(s => (
          <div key={s.id} className="df-item-row">
            <span className="field-input" style={{ cursor: 'default' }}>{s.prename} {s.name}</span>
            <button type="button" className="df-remove-btn" onClick={() => removeStudent(s.id)}>{t.domain.common.remove}</button>
          </div>
        ))}
        <div className="df-item-row">
          <select
            className="field-input field-select"
            value={studentSelect}
            onChange={e => setStudentSelect(e.target.value)}
          >
            <option value="">{t.domain.common.select}</option>
            {studentsList
              .filter(s => !students.find(as => as.id === s.id))
              .map(s => (
                <option key={s.id} value={s.id}>{s.prename} {s.name}</option>
              ))}
          </select>
          <button type="button" className="df-add-btn" onClick={addStudent}>{t.domain.common.add}</button>
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
