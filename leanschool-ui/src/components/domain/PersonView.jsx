import { useState, useEffect } from 'react';
import { useAuth } from '../../auth/useAuth';
import { useTranslation } from '../../i18n/useTranslation';
import './domain-components.css';
import { config } from '../../config';

const API = config.leanschoolUrl;

export default function PersonView({ id }) {
  const { t } = useTranslation();
  const { authFetch } = useAuth();
  const [person, setPerson] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    if (!id) return;
    setLoading(true);
    setError(null);
    authFetch(`${API}/persons/${id}`)
      .then(res => {
        if (!res.ok) throw new Error(`Failed to load person (${res.status})`);
        return res.json();
      })
      .then(data => setPerson(data))
      .catch(err => setError(err.message))
      .finally(() => setLoading(false));
  }, [id]);

  if (loading) {
    return (
      <div className="dv-loading">
        <div className="spinner" />
      </div>
    );
  }

  if (error) return <div className="dv-error">{error}</div>;
  if (!person) return null;

  const dob = person.dateOfBirth
    ? new Date(person.dateOfBirth).toLocaleDateString()
    : '—';

  return (
    <div className="dv-card">
      <div className="dv-row">
        <span className="dv-label">{t.domain.common.id}</span>
        <span className="dv-value">{person.id}</span>
      </div>
      <div className="dv-row">
        <span className="dv-label">{t.domain.person.prename}</span>
        <span className="dv-value">{person.prename}</span>
      </div>
      <div className="dv-row">
        <span className="dv-label">{t.domain.person.name}</span>
        <span className="dv-value">{person.name}</span>
      </div>
      <div className="dv-row">
        <span className="dv-label">{t.domain.person.dateOfBirth}</span>
        <span className="dv-value">{dob}</span>
      </div>
      {person.sub && (
        <div className="dv-row">
          <span className="dv-label">{t.domain.person.sub}</span>
          <span className="dv-value">{person.sub}</span>
        </div>
      )}
      {person.address && (
        <>
          <div className="dv-section-title">{t.domain.person.address}</div>
          <div className="dv-row">
            <span className="dv-label">{t.domain.address.street}</span>
            <span className="dv-value">
              {person.address.street} {person.address.number}
            </span>
          </div>
          <div className="dv-row">
            <span className="dv-label">{t.domain.address.postalCode}</span>
            <span className="dv-value">{person.address.postalCode}</span>
          </div>
        </>
      )}
      <div className="dv-row">
        <span className="dv-label">{t.domain.common.version}</span>
        <span className="dv-value">{person.version}</span>
      </div>
    </div>
  );
}
