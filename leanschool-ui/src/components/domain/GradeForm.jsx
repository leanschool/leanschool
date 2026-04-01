import { useState, useEffect } from 'react';
import { useAuth } from '../../auth/useAuth';
import { useTranslation } from '../../i18n/useTranslation';
import './domain-components.css';
import { config } from '../../config';

const API = config.leanschoolUrl;

export default function GradeForm({ id, persist = false, onSave, onCancel }) {
  const { t } = useTranslation();
  const { authFetch } = useAuth();
  const [gradeValue, setGradeValue] = useState('');
  const [examId, setExamId] = useState('');
  const [studentId, setStudentId] = useState('');
  const [version, setVersion] = useState(0);
  const [exams, setExams] = useState([]);
  const [students, setStudents] = useState([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState(null);
  const [conflict, setConflict] = useState(false);

  useEffect(() => {
    const fetches = [
      authFetch(`${API}/exams`).then(res => res.ok ? res.json() : []).catch(() => []),
      authFetch(`${API}/students`).then(res => res.ok ? res.json() : []).catch(() => []),
    ];

    if (id) {
      fetches.push(
        authFetch(`${API}/grades/${id}`)
          .then(res => {
            if (!res.ok) throw new Error(`Failed to load grade (${res.status})`);
            return res.json();
          })
      );
    }

    Promise.all(fetches)
      .then(([examList, studentList, grade]) => {
        setExams(examList || []);
        setStudents(studentList || []);
        if (grade) {
          setGradeValue(grade.grade != null ? String(grade.grade) : '');
          setExamId(grade.exam?.id ?? '');
          setStudentId(grade.student?.id ?? '');
          setVersion(grade.version);
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
      grade: parseFloat(gradeValue),
      ...(examId ? { exam: { id: examId } } : {}),
      ...(studentId ? { student: { id: studentId } } : {}),
      ...(id ? { version } : {}),
    };

    if (!persist) {
      onSave?.(body);
      return;
    }

    setSaving(true);
    const url = id ? `${API}/grades/${id}` : `${API}/grades`;
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
          <div className="df-label">{t.domain.grade.grade}</div>
          <input
            className="field-input"
            type="number"
            step="0.5"
            min="1"
            max="6"
            value={gradeValue}
            onChange={e => setGradeValue(e.target.value)}
            required
          />
        </div>
        <div className="df-field-group">
          <div className="df-label">{t.domain.grade.exam}</div>
          <select
            className="field-input field-select"
            value={examId}
            onChange={e => setExamId(e.target.value)}
          >
            <option value="">{t.domain.common.none}</option>
            {exams.map(ex => (
              <option key={ex.id} value={ex.id}>
                {ex.id}
              </option>
            ))}
          </select>
        </div>
        <div className="df-field-group">
          <div className="df-label">{t.domain.grade.student}</div>
          <select
            className="field-input field-select"
            value={studentId}
            onChange={e => setStudentId(e.target.value)}
          >
            <option value="">{t.domain.common.none}</option>
            {students.map(s => (
              <option key={s.id} value={s.id}>
                {s.prename} {s.name}
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
