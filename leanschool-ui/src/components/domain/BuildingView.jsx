import { useState, useEffect } from 'react';
import { useAuth } from '../../auth/useAuth';
import { useTranslation } from '../../i18n/useTranslation';
import './domain-components.css';

const API = import.meta.env.VITE_LEANSCHOOL_URL || 'http://localhost:8080';

export default function BuildingView({ id }) {
  const { t } = useTranslation();
  const { authFetch } = useAuth();
  const [building, setBuilding] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    if (!id) return;
    setLoading(true);
    setError(null);
    authFetch(`${API}/buildings/${id}`)
      .then(res => {
        if (!res.ok) throw new Error(t.domain.common.loadError);
        return res.json();
      })
      .then(data => setBuilding(data))
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
  if (!building) return null;

  return (
    <div className="dv-card">
      <div className="dv-row">
        <span className="dv-label">{t.domain.common.id}</span>
        <span className="dv-value">{building.id}</span>
      </div>
      <div className="dv-row">
        <span className="dv-label">{t.domain.building.name}</span>
        <span className="dv-value">{building.name}</span>
      </div>
      <div className="dv-row">
        <span className="dv-label">{t.domain.common.version}</span>
        <span className="dv-value">{building.version}</span>
      </div>

      {building.location && (
        <>
          <div className="dv-section-title">{t.domain.building.location}</div>
          <div className="dv-row">
            <span className="dv-label">{t.domain.location.lon}</span>
            <span className="dv-value">{building.location.lon}</span>
          </div>
          <div className="dv-row">
            <span className="dv-label">{t.domain.location.lat}</span>
            <span className="dv-value">{building.location.lat}</span>
          </div>
        </>
      )}
    </div>
  );
}
