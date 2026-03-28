import { useState, useEffect } from 'react';
import { useAuth } from '../../auth/useAuth';
import { useTranslation } from '../../i18n/useTranslation';
import './domain-components.css';

const API = import.meta.env.VITE_LEANSCHOOL_URL || 'http://localhost:8080';

export default function CurriculumView({ id }) {
  const { t } = useTranslation();
  const { authFetch } = useAuth();
  const [curriculum, setCurriculum] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    if (!id) return;
    setLoading(true);
    setError(null);
    authFetch(`${API}/curricula/${id}`)
      .then(res => {
        if (!res.ok) throw new Error(`Failed to load curriculum (${res.status})`);
        return res.json();
      })
      .then(data => setCurriculum(data))
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
  if (!curriculum) return null;

  return (
    <div className="dv-card">
      <div className="dv-row">
        <span className="dv-label">{t.domain.common.id}</span>
        <span className="dv-value">{curriculum.id}</span>
      </div>
      {curriculum.name && (
        <div className="dv-row">
          <span className="dv-label">{t.domain.curriculum.name}</span>
          <span className="dv-value">{curriculum.name}</span>
        </div>
      )}
      {curriculum.activeSince && (
        <div className="dv-row">
          <span className="dv-label">{t.domain.curriculum.activeSince}</span>
          <span className="dv-value">{new Date(curriculum.activeSince).toLocaleDateString()}</span>
        </div>
      )}
      {curriculum.activeUntil && (
        <div className="dv-row">
          <span className="dv-label">{t.domain.curriculum.activeUntil}</span>
          <span className="dv-value">{new Date(curriculum.activeUntil).toLocaleDateString()}</span>
        </div>
      )}
      {curriculum.activeFrom && (
        <div className="dv-row">
          <span className="dv-label">{t.domain.curriculum.activeFrom}</span>
          <span className="dv-value">{new Date(curriculum.activeFrom).toLocaleDateString()}</span>
        </div>
      )}
      <div className="dv-row">
        <span className="dv-label">{t.domain.common.version}</span>
        <span className="dv-value">{curriculum.version}</span>
      </div>
    </div>
  );
}
