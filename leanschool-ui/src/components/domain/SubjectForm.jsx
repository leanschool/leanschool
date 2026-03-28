import { useState, useEffect } from 'react';
import { useAuth } from '../../auth/useAuth';
import { useTranslation } from '../../i18n/useTranslation';
import './domain-components.css';

const API = import.meta.env.VITE_LEANSCHOOL_URL || 'http://localhost:8080';

export default function SubjectForm({ id, persist = false, onSave, onCancel }) {
  const { t } = useTranslation();
  const { authFetch } = useAuth();
  const [name, setName] = useState('');
  const [teachers, setTeachers] = useState([]);
  const [allTeachers, setAllTeachers] = useState([]);
  const [selectedTeacherId, setSelectedTeacherId] = useState('');
  const [version, setVersion] = useState(0);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState(null);
  const [conflict, setConflict] = useState(false);

  useEffect(() => {
    const fetches = [
      authFetch(`${API}/teachers`)
        .then(res => res.ok ? res.json() : [])
        .catch(() => []),
    ];

    if (id) {
      fetches.push(
        authFetch(`${API}/subjects/${id}`)
          .then(res => {
            if (!res.ok) throw new Error(`Failed to load subject (${res.status})`);
            return res.json();
          })
      );
    }

    Promise.all(fetches)
      .then(([teacherList, subject]) => {
        setAllTeachers(teacherList || []);
        if (subject) {
          setName(subject.name ?? '');
          setTeachers(subject.teachers || []);
          setVersion(subject.version);
        }
      })
      .catch(err => setError(err.message))
      .finally(() => setLoading(false));
  }, [id]);

  function addTeacher() {
    if (!selectedTeacherId) return;
    const teacher = allTeachers.find(t => String(t.id) === String(selectedTeacherId));
    if (!teacher) return;
    if (teachers.some(t => String(t.id) === String(selectedTeacherId))) return;
    setTeachers(prev => [...prev, teacher]);
    setSelectedTeacherId('');
  }

  function removeTeacher(teacherId) {
    setTeachers(prev => prev.filter(t => String(t.id) !== String(teacherId)));
  }

  async function handleSubmit(e) {
    e.preventDefault();
    setConflict(false);
    setError(null);

    const body = {
      ...(name ? { name } : {}),
      teachers: teachers.map(t => ({ id: t.id })),
      ...(id ? { version } : {}),
    };

    if (!persist) {
      onSave?.(body);
      return;
    }

    setSaving(true);
    const url = id ? `${API}/subjects/${id}` : `${API}/subjects`;
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
          <div className="df-label">{t.domain.subject.name}</div>
          <input
            className="field-input"
            type="text"
            value={name}
            onChange={e => setName(e.target.value)}
          />
        </div>
      </div>

      <div className="df-section">
        <div className="df-label">{t.domain.subject.teachers}</div>
        {teachers.map(teacher => (
          <div key={teacher.id} className="df-item-row">
            <span>{teacher.prename} {teacher.name}</span>
            <button
              type="button"
              className="df-remove-btn"
              onClick={() => removeTeacher(teacher.id)}
            >
              {t.domain.common.remove}
            </button>
          </div>
        ))}
        <div className="df-item-row">
          <select
            className="field-input field-select"
            value={selectedTeacherId}
            onChange={e => setSelectedTeacherId(e.target.value)}
          >
            <option value="">{t.domain.common.select}</option>
            {allTeachers
              .filter(t => !teachers.some(sel => String(sel.id) === String(t.id)))
              .map(t => (
                <option key={t.id} value={t.id}>
                  {t.prename} {t.name}
                </option>
              ))}
          </select>
          <button type="button" className="df-add-btn" onClick={addTeacher}>
            {t.domain.common.add}
          </button>
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
