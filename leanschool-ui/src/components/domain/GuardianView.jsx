import { useState, useEffect } from 'react';
import { useAuth } from '../../auth/useAuth';
import { useTranslation } from '../../i18n/useTranslation';
import './domain-components.css';
import { config } from '../../config';

const API = config.leanschoolUrl;

export default function GuardianView({ id }) {
  const { t } = useTranslation();
  const { authFetch } = useAuth();
  const [guardian, setGuardian] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    if (!id) return;
    setLoading(true);
    setError(null);
    authFetch(`${API}/guardians/${id}`)
      .then(res => {
        if (!res.ok) throw new Error(`Failed to load guardian (${res.status})`);
        return res.json();
      })
      .then(data => setGuardian(data))
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
  if (!guardian) return null;

  const dob = guardian.dateOfBirth
    ? new Date(guardian.dateOfBirth).toLocaleDateString()
    : '—';

  return (
    <div className="dv-card">
      <div className="dv-row">
        <span className="dv-label">{t.domain.common.id}</span>
        <span className="dv-value">{guardian.id}</span>
      </div>
      <div className="dv-row">
        <span className="dv-label">{t.domain.person.prename}</span>
        <span className="dv-value">{guardian.prename}</span>
      </div>
      <div className="dv-row">
        <span className="dv-label">{t.domain.person.name}</span>
        <span className="dv-value">{guardian.name}</span>
      </div>
      <div className="dv-row">
        <span className="dv-label">{t.domain.person.dateOfBirth}</span>
        <span className="dv-value">{dob}</span>
      </div>
      {guardian.sub && (
        <div className="dv-row">
          <span className="dv-label">{t.domain.person.sub}</span>
          <span className="dv-value">{guardian.sub}</span>
        </div>
      )}
      {guardian.address && (
        <>
          <div className="dv-section-title">{t.domain.person.address}</div>
          <div className="dv-row">
            <span className="dv-label">{t.domain.address.street}</span>
            <span className="dv-value">
              {guardian.address.street} {guardian.address.number}
            </span>
          </div>
          <div className="dv-row">
            <span className="dv-label">{t.domain.address.postalCode}</span>
            <span className="dv-value">{guardian.address.postalCode}</span>
          </div>
        </>
      )}
      <div className="dv-row">
        <span className="dv-label">{t.domain.common.version}</span>
        <span className="dv-value">{guardian.version}</span>
      </div>
    </div>
  );
}
