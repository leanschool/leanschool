import { useState, useEffect } from 'react';
import { useAuth } from '../../auth/useAuth';
import { useTranslation } from '../../i18n/useTranslation';
import './domain-components.css';
import { config } from '../../config';

const API = config.leanschoolUrl;

export default function LocationView({ id }) {
  const { t } = useTranslation();
  const { authFetch } = useAuth();
  const [location, setLocation] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    if (!id) return;
    authFetch(`${API}/locations/${id}`)
      .then(res => {
        if (!res.ok) throw new Error(t.domain.common.loadError);
        return res.json();
      })
      .then(data => setLocation(data))
      .catch(err => setError(err.message))
      .finally(() => setLoading(false));
  }, [authFetch, id, t.domain.common.loadError]);

  if (loading) {
    return (
      <div className="dv-loading">
        <div className="spinner" />
      </div>
    );
  }

  if (error) return <div className="dv-error">{error}</div>;
  if (!location) return null;

  return (
    <div className="dv-card">
      <div className="dv-row">
        <span className="dv-label">{t.domain.common.id}</span>
        <span className="dv-value">{location.id}</span>
      </div>
      <div className="dv-row">
        <span className="dv-label">{t.domain.location.lon}</span>
        <span className="dv-value">{location.lon}</span>
      </div>
      <div className="dv-row">
        <span className="dv-label">{t.domain.location.lat}</span>
        <span className="dv-value">{location.lat}</span>
      </div>
      <div className="dv-row">
        <span className="dv-label">{t.domain.common.version}</span>
        <span className="dv-value">{location.version}</span>
      </div>
    </div>
  );
}
