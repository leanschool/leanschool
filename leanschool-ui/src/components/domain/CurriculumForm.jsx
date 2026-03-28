import { useState, useEffect } from 'react';
import { useAuth } from '../../auth/useAuth';
import { useTranslation } from '../../i18n/useTranslation';
import './domain-components.css';

const API = import.meta.env.VITE_LEANSCHOOL_URL || 'http://localhost:8080';

export default function CurriculumForm({ id, persist = false, onSave, onCancel }) {
  const { t } = useTranslation();
  const { authFetch } = useAuth();
  const [fields, setFields] = useState({ name: '', activeSince: '', activeUntil: '', activeFrom: '' });
  const [version, setVersion] = useState(0);
  const [loading, setLoading] = useState(!!id);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState(null);
  const [conflict, setConflict] = useState(false);

  useEffect(() => {
    if (!id) return;
    setLoading(true);
    authFetch(`${API}/curricula/${id}`)
      .then(res => {
        if (!res.ok) throw new Error(`Failed to load curriculum (${res.status})`);
        return res.json();
      })
      .then(data => {
        setFields({
          name: data.name ?? '',
          activeSince: data.activeSince ? data.activeSince.slice(0, 10) : '',
          activeUntil: data.activeUntil ? data.activeUntil.slice(0, 10) : '',
          activeFrom: data.activeFrom ? data.activeFrom.slice(0, 10) : '',
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
      ...(fields.name ? { name: fields.name } : {}),
      ...(fields.activeSince ? { activeSince: fields.activeSince } : {}),
      ...(fields.activeUntil ? { activeUntil: fields.activeUntil } : {}),
      ...(fields.activeFrom ? { activeFrom: fields.activeFrom } : {}),
      ...(id ? { version } : {}),
    };

    if (!persist) {
      onSave?.(body);
      return;
    }

    setSaving(true);
    const url = id ? `${API}/curricula/${id}` : `${API}/curricula`;
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
          <div className="df-label">{t.domain.curriculum.name}</div>
          <input
            className="field-input"
            type="text"
            name="name"
            value={fields.name}
            onChange={handleChange}
          />
        </div>
        <div className="df-field-group">
          <div className="df-label">{t.domain.curriculum.activeSince}</div>
          <input
            className="field-input"
            type="date"
            name="activeSince"
            value={fields.activeSince}
            onChange={handleChange}
          />
        </div>
        <div className="df-field-group">
          <div className="df-label">{t.domain.curriculum.activeUntil}</div>
          <input
            className="field-input"
            type="date"
            name="activeUntil"
            value={fields.activeUntil}
            onChange={handleChange}
          />
        </div>
        <div className="df-field-group">
          <div className="df-label">{t.domain.curriculum.activeFrom}</div>
          <input
            className="field-input"
            type="date"
            name="activeFrom"
            value={fields.activeFrom}
            onChange={handleChange}
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
