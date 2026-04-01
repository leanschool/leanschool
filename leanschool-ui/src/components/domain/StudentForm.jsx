import { useState, useEffect } from 'react';
import { useAuth } from '../../auth/useAuth';
import { useTranslation } from '../../i18n/useTranslation';
import './domain-components.css';
import { config } from '../../config';

const API = config.leanschoolUrl;

export default function StudentForm({ id, persist = false, onSave, onCancel }) {
  const { t } = useTranslation();
  const { authFetch } = useAuth();
  const [fields, setFields] = useState({ name: '', prename: '', dateOfBirth: '' });
  const [version, setVersion] = useState(0);
  const [loading, setLoading] = useState(!!id);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState(null);
  const [conflict, setConflict] = useState(false);

  const [addresses, setAddresses] = useState([]);
  const [selectedAddress, setSelectedAddress] = useState('');
  const [schoolClasses, setSchoolClasses] = useState([]);
  const [currentSchoolClass, setCurrentSchoolClass] = useState('');
  const [guardiansList, setGuardiansList] = useState([]);
  const [guardians, setGuardians] = useState([]);
  const [guardianSelect, setGuardianSelect] = useState('');
  const [pastSchoolClasses, setPastSchoolClasses] = useState([]);
  const [pastClassSelect, setPastClassSelect] = useState('');

  useEffect(() => {
    authFetch(`${API}/addresses`)
      .then(res => res.ok ? res.json() : [])
      .then(data => setAddresses(Array.isArray(data) ? data : []))
      .catch(() => {});

    authFetch(`${API}/school-classes`)
      .then(res => res.ok ? res.json() : [])
      .then(data => setSchoolClasses(Array.isArray(data) ? data : []))
      .catch(() => {});

    authFetch(`${API}/guardians`)
      .then(res => res.ok ? res.json() : [])
      .then(data => setGuardiansList(Array.isArray(data) ? data : []))
      .catch(() => {});
  }, []);

  useEffect(() => {
    if (!id) return;
    setLoading(true);
    authFetch(`${API}/students/${id}`)
      .then(res => {
        if (!res.ok) throw new Error(`Failed to load student (${res.status})`);
        return res.json();
      })
      .then(data => {
        setFields({
          name: data.name || '',
          prename: data.prename || '',
          dateOfBirth: data.dateOfBirth ? data.dateOfBirth.slice(0, 10) : '',
        });
        setVersion(data.version);
        setSelectedAddress(data.address?.id || '');
        setCurrentSchoolClass(data.currentSchoolClass?.id || '');
        setGuardians(data.guardians || []);
        setPastSchoolClasses(data.pastSchoolClasses || []);
      })
      .catch(err => setError(err.message))
      .finally(() => setLoading(false));
  }, [id]);

  function handleChange(e) {
    const { name, value } = e.target;
    setFields(prev => ({ ...prev, [name]: value }));
  }

  function addGuardian() {
    if (!guardianSelect) return;
    const guardian = guardiansList.find(g => String(g.id) === String(guardianSelect));
    if (!guardian || guardians.find(g => g.id === guardian.id)) return;
    setGuardians(prev => [...prev, guardian]);
    setGuardianSelect('');
  }

  function removeGuardian(gid) {
    setGuardians(prev => prev.filter(g => g.id !== gid));
  }

  function addPastClass() {
    if (!pastClassSelect) return;
    const sc = schoolClasses.find(c => String(c.id) === String(pastClassSelect));
    if (!sc || pastSchoolClasses.find(c => c.id === sc.id)) return;
    setPastSchoolClasses(prev => [...prev, sc]);
    setPastClassSelect('');
  }

  function removePastClass(cid) {
    setPastSchoolClasses(prev => prev.filter(c => c.id !== cid));
  }

  async function handleSubmit(e) {
    e.preventDefault();
    setConflict(false);
    setError(null);

    const body = {
      name: fields.name,
      prename: fields.prename,
      ...(fields.dateOfBirth ? { dateOfBirth: fields.dateOfBirth } : {}),
      ...(selectedAddress ? { address: { id: selectedAddress } } : {}),
      ...(currentSchoolClass ? { currentSchoolClass: { id: currentSchoolClass } } : {}),
      guardians: guardians.map(g => ({ id: g.id })),
      pastSchoolClasses: pastSchoolClasses.map(c => ({ id: c.id })),
      ...(id ? { version } : {}),
    };

    if (!persist) {
      onSave?.(body);
      return;
    }

    setSaving(true);
    const url = id ? `${API}/students/${id}` : `${API}/students`;
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
          <div className="df-label">{t.domain.person.address}</div>
          <select
            className="field-input field-select"
            value={selectedAddress}
            onChange={e => setSelectedAddress(e.target.value)}
          >
            <option value="">{t.domain.common.none}</option>
            {addresses.map(a => (
              <option key={a.id} value={a.id}>{a.name || a.id}</option>
            ))}
          </select>
        </div>
        <div className="df-field-group">
          <div className="df-label">{t.domain.student.currentSchoolClass}</div>
          <select
            className="field-input field-select"
            value={currentSchoolClass}
            onChange={e => setCurrentSchoolClass(e.target.value)}
          >
            <option value="">{t.domain.common.none}</option>
            {schoolClasses.map(c => (
              <option key={c.id} value={c.id}>{c.name}</option>
            ))}
          </select>
        </div>
      </div>

      <div className="df-section">
        <div className="df-section-title">{t.domain.student.guardians}</div>
        {guardians.map(g => (
          <div key={g.id} className="df-item-row">
            <span className="field-input" style={{ cursor: 'default' }}>{g.prename} {g.name}</span>
            <button type="button" className="df-remove-btn" onClick={() => removeGuardian(g.id)}>✕</button>
          </div>
        ))}
        <div className="df-item-row">
          <select
            className="field-input field-select"
            value={guardianSelect}
            onChange={e => setGuardianSelect(e.target.value)}
          >
            <option value="">{t.domain.common.select}</option>
            {guardiansList
              .filter(g => !guardians.find(ag => ag.id === g.id))
              .map(g => (
                <option key={g.id} value={g.id}>{g.prename} {g.name}</option>
              ))}
          </select>
          <button type="button" className="df-add-btn" onClick={addGuardian}>{t.domain.common.add}</button>
        </div>
      </div>

      <div className="df-section">
        <div className="df-section-title">{t.domain.student.pastSchoolClasses}</div>
        {pastSchoolClasses.map(c => (
          <div key={c.id} className="df-item-row">
            <span className="field-input" style={{ cursor: 'default' }}>{c.name}</span>
            <button type="button" className="df-remove-btn" onClick={() => removePastClass(c.id)}>✕</button>
          </div>
        ))}
        <div className="df-item-row">
          <select
            className="field-input field-select"
            value={pastClassSelect}
            onChange={e => setPastClassSelect(e.target.value)}
          >
            <option value="">{t.domain.common.select}</option>
            {schoolClasses
              .filter(c => !pastSchoolClasses.find(pc => pc.id === c.id))
              .map(c => (
                <option key={c.id} value={c.id}>{c.name}</option>
              ))}
          </select>
          <button type="button" className="df-add-btn" onClick={addPastClass}>{t.domain.common.add}</button>
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
