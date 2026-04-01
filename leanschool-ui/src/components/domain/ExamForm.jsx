import { useState, useEffect } from 'react';
import { useAuth } from '../../auth/useAuth';
import { useTranslation } from '../../i18n/useTranslation';
import './domain-components.css';
import { config } from '../../config';

const API = config.leanschoolUrl;

export default function ExamForm({ id, persist = false, onSave, onCancel }) {
  const { t } = useTranslation();
  const { authFetch } = useAuth();
  const [schoolClassId, setSchoolClassId] = useState('');
  const [version, setVersion] = useState(0);
  const [schoolClasses, setSchoolClasses] = useState([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState(null);
  const [conflict, setConflict] = useState(false);

  useEffect(() => {
    const fetches = [
      authFetch(`${API}/school-classes`).then(res => res.ok ? res.json() : []).catch(() => []),
    ];

    if (id) {
      fetches.push(
        authFetch(`${API}/exams/${id}`)
          .then(res => {
            if (!res.ok) throw new Error(`Failed to load exam (${res.status})`);
            return res.json();
          })
      );
    }

    Promise.all(fetches)
      .then(([classList, exam]) => {
        setSchoolClasses(classList || []);
        if (exam) {
          setSchoolClassId(exam.schoolClass?.id ?? '');
          setVersion(exam.version);
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
      ...(schoolClassId ? { schoolClass: { id: schoolClassId } } : {}),
      ...(id ? { version } : {}),
    };

    if (!persist) {
      onSave?.(body);
      return;
    }

    setSaving(true);
    const url = id ? `${API}/exams/${id}` : `${API}/exams`;
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
          <div className="df-label">{t.domain.exam.schoolClass}</div>
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
