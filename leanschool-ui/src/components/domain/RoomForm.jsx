import { useState, useEffect } from 'react';
import { useAuth } from '../../auth/useAuth';
import { useTranslation } from '../../i18n/useTranslation';
import './domain-components.css';

const API = import.meta.env.VITE_LEANSCHOOL_URL || 'http://localhost:8080';

export default function RoomForm({ id, persist = false, onSave, onCancel }) {
  const { t } = useTranslation();
  const { authFetch } = useAuth();
  const [fields, setFields] = useState({ name: '', roomType: '' });
  const [buildingId, setBuildingId] = useState('');
  const [version, setVersion] = useState(0);
  const [buildings, setBuildings] = useState([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState(null);
  const [conflict, setConflict] = useState(false);

  useEffect(() => {
    const fetches = [
      authFetch(`${API}/buildings`)
        .then(res => res.ok ? res.json() : [])
        .catch(() => []),
    ];

    if (id) {
      fetches.push(
        authFetch(`${API}/rooms/${id}`)
          .then(res => {
            if (!res.ok) throw new Error(t.domain.common.loadError);
            return res.json();
          })
      );
    }

    Promise.all(fetches)
      .then(([bldgs, room]) => {
        setBuildings(bldgs || []);
        if (room) {
          setFields({ name: room.name, roomType: room.roomType || '' });
          setBuildingId(room.building?.id ?? '');
          setVersion(room.version);
        }
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
      name: fields.name,
      building: { id: buildingId },
      ...(fields.roomType ? { roomType: fields.roomType } : {}),
      ...(id ? { version } : {}),
    };

    if (!persist) {
      onSave?.(body);
      return;
    }

    setSaving(true);
    const url = id ? `${API}/rooms/${id}` : `${API}/rooms`;
    const method = id ? 'PUT' : 'POST';

    try {
      const res = await authFetch(url, {
        method,
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body),
      });

      if (res.status === 409) {
        setConflict(true);
        return;
      }
      if (!res.ok) throw new Error(t.domain.common.saveError);

      const saved = await res.json();
      onSave?.(saved);
    } catch (err) {
      setError(err.message);
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
          <div className="df-label">{t.domain.room.name}</div>
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
          <div className="df-label">{t.domain.room.building}</div>
          <select
            className="field-input field-select"
            value={buildingId}
            onChange={e => setBuildingId(e.target.value)}
            required
          >
            <option value="" disabled>{t.domain.common.select}</option>
            {buildings.map(b => (
              <option key={b.id} value={b.id}>
                {b.name}
              </option>
            ))}
          </select>
        </div>
        <div className="df-field-group">
          <div className="df-label">{t.domain.room.roomType}</div>
          <select
            className="field-input field-select"
            name="roomType"
            value={fields.roomType}
            onChange={handleChange}
          >
            <option value="">{t.domain.common.none}</option>
            {Object.entries(t.domain.roomTypes).map(([key, label]) => (
              <option key={key} value={key}>
                {label}
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
