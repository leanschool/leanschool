import { useState, useEffect } from 'react';
import { useAuth } from '../../auth/useAuth';
import { useTranslation } from '../../i18n/useTranslation';
import './domain-components.css';
import { config } from '../../config';

const API = config.leanschoolUrl;

export default function SubjectView({ id }) {
  const { t } = useTranslation();
  const { authFetch } = useAuth();
  const [subject, setSubject] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    if (!id) return;
    setLoading(true);
    setError(null);
    authFetch(`${API}/subjects/${id}`)
      .then(res => {
        if (!res.ok) throw new Error(`Failed to load subject (${res.status})`);
        return res.json();
      })
      .then(data => setSubject(data))
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
  if (!subject) return null;

  return (
    <div className="dv-card">
      <div className="dv-row">
        <span className="dv-label">{t.domain.common.id}</span>
        <span className="dv-value">{subject.id}</span>
      </div>
      {subject.name && (
        <div className="dv-row">
          <span className="dv-label">{t.domain.subject.name}</span>
          <span className="dv-value">{subject.name}</span>
        </div>
      )}
      {subject.teachers && subject.teachers.length > 0 && (
        <>
          <div className="dv-section-title">{t.domain.subject.teachers}</div>
          <div className="dv-badge-list">
            {subject.teachers.map(t => (
              <span key={t.id} className="dv-badge">
                {t.prename} {t.name}
              </span>
            ))}
          </div>
        </>
      )}
      <div className="dv-row">
        <span className="dv-label">{t.domain.common.version}</span>
        <span className="dv-value">{subject.version}</span>
      </div>
    </div>
  );
}
