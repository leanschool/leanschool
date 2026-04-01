import { useState, useEffect } from 'react';
import { useAuth } from '../../auth/useAuth';
import { useTranslation } from '../../i18n/useTranslation';
import './domain-components.css';
import { config } from '../../config';

const API = config.leanschoolUrl;

const EMPTY = {
  name: '',
  prename: '',
  dateOfBirth: '',
  sub: '',
  addressId: '',
  iban: '',
  atSchoolSince: '',
  atSchoolUntil: '',
};

export default function TeacherForm({ id, persist = false, onSave, onCancel }) {
  const { t } = useTranslation();
  const { authFetch } = useAuth();
  const [fields, setFields] = useState(EMPTY);
  const [version, setVersion] = useState(0);
  const [addresses, setAddresses] = useState([]);
  const [loading, setLoading] = useState(!!id);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState(null);
  const [conflict, setConflict] = useState(false);

  useEffect(() => {
    authFetch(`${API}/addresses`)
      .then(res => res.ok ? res.json() : [])
      .then(data => setAddresses(data))
      .catch(() => {});
  }, []);

  useEffect(() => {
    if (!id) return;
    setLoading(true);
    authFetch(`${API}/teachers/${id}`)
      .then(res => {
        if (!res.ok) throw new Error(`Failed to load teacher (${res.status})`);
        return res.json();
      })
      .then(data => {
        setFields({
          name: data.name ?? '',
          prename: data.prename ?? '',
          dateOfBirth: data.dateOfBirth ? data.dateOfBirth.slice(0, 10) : '',
          sub: data.sub ?? '',
          addressId: data.address?.id ?? '',
          iban: data.iban ?? '',
          atSchoolSince: data.atSchoolSince ? data.atSchoolSince.slice(0, 10) : '',
          atSchoolUntil: data.atSchoolUntil ? data.atSchoolUntil.slice(0, 10) : '',
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
      name: fields.name,
      prename: fields.prename,
      atSchoolSince: fields.atSchoolSince + 'T00:00:00Z',
      ...(fields.dateOfBirth ? { dateOfBirth: fields.dateOfBirth + 'T00:00:00Z' } : {}),
      ...(fields.sub ? { sub: fields.sub } : {}),
      ...(fields.iban ? { iban: fields.iban } : {}),
      ...(fields.atSchoolUntil ? { atSchoolUntil: fields.atSchoolUntil + 'T00:00:00Z' } : {}),
      ...(fields.addressId ? { address: { id: fields.addressId } } : {}),
      ...(id ? { version } : {}),
    };

    if (!persist) {
      onSave?.(body);
      return;
    }

    setSaving(true);
    const url = id ? `${API}/teachers/${id}` : `${API}/teachers`;
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
      if (!res.ok) throw new Error(`Save failed (${res.status})`);

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
          <div className="df-label">{t.domain.person.prename}</div>
          <input
            className="field-input"
            type="text"
            name="prename"
            value={fields.prename}
            onChange={handleChange}
            required
          />
        </div>
        <div className="df-field-group">
          <div className="df-label">{t.domain.person.name}</div>
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
          <div className="df-label">{t.domain.person.dateOfBirth}</div>
          <input
            className="field-input"
            type="date"
            name="dateOfBirth"
            value={fields.dateOfBirth}
            onChange={handleChange}
          />
        </div>
        <div className="df-field-group">
          <div className="df-label">{t.domain.person.sub}</div>
          <input
            className="field-input"
            type="text"
            name="sub"
            value={fields.sub}
            onChange={handleChange}
          />
        </div>
        <div className="df-field-group">
          <div className="df-label">{t.domain.teacher.iban}</div>
          <input
            className="field-input"
            type="text"
            name="iban"
            value={fields.iban}
            onChange={handleChange}
          />
        </div>
        <div className="df-field-group">
          <div className="df-label">{t.domain.teacher.atSchoolSince}</div>
          <input
            className="field-input"
            type="date"
            name="atSchoolSince"
            value={fields.atSchoolSince}
            onChange={handleChange}
            required
          />
        </div>
        <div className="df-field-group">
          <div className="df-label">{t.domain.teacher.atSchoolUntil}</div>
          <input
            className="field-input"
            type="date"
            name="atSchoolUntil"
            value={fields.atSchoolUntil}
            onChange={handleChange}
          />
        </div>
        <div className="df-field-group">
          <div className="df-label">{t.domain.person.address}</div>
          <select
            className="field-input field-select"
            name="addressId"
            value={fields.addressId}
            onChange={handleChange}
          >
            <option value="">{t.domain.common.none}</option>
            {addresses.map(a => (
              <option key={a.id} value={a.id}>
                {a.street} {a.number}
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
