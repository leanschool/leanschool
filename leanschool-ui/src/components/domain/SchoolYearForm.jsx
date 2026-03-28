import { useState, useEffect } from 'react';
import { useAuth } from '../../auth/useAuth';
import { useTranslation } from '../../i18n/useTranslation';
import './domain-components.css';

const API = import.meta.env.VITE_LEANSCHOOL_URL || 'http://localhost:8080';

export default function SchoolYearForm({ id, persist = false, onSave, onCancel }) {
  const { t } = useTranslation();
  const { authFetch } = useAuth();
  const [fields, setFields] = useState({ from: '', to: '' });
  const [version, setVersion] = useState(0);
  const [loading, setLoading] = useState(!!id);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState(null);
  const [conflict, setConflict] = useState(false);

  useEffect(() => {
    if (!id) return;
    setLoading(true);
    authFetch(`${API}/school-years/${id}`)
      .then(res => {
        if (!res.ok) throw new Error(`Failed to load school year (${res.status})`);
        return res.json();
      })
      .then(data => {
        setFields({
          from: data.from ? data.from.slice(0, 10) : '',
          to: data.to ? data.to.slice(0, 10) : '',
        });
        setVersion(data.version);
      })
      .catch(err => setError(err.message))
      .finally(() => setLoading(false));
  }, [id]);

  function handleChange(e) {
    const { name, value } = e.target;
    setFields(prev => ({ ...prev, [name]: value }));
  }

  async function handleSubmit(e) {
    e.preventDefault();
    setConflict(false);
    setError(null);

    const body = {
      from: fields.from ? `${fields.from}T00:00:00Z` : fields.from,
      to: fields.to ? `${fields.to}T00:00:00Z` : fields.to,
      ...(id ? { version } : {}),
    };

    if (!persist) {
      onSave?.(body);
      return;
    }

    setSaving(true);
    const url = id ? `${API}/school-years/${id}` : `${API}/school-years`;
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
          <div className="df-label">{t.domain.schoolYear.from}</div>
          <input
            className="field-input"
            type="date"
            name="from"
            value={fields.from}
            onChange={handleChange}
            required
          />
        </div>
        <div className="df-field-group">
          <div className="df-label">{t.domain.schoolYear.to}</div>
          <input
            className="field-input"
            type="date"
            name="to"
            value={fields.to}
            onChange={handleChange}
            required
          />
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
