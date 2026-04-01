import { useState, useEffect } from 'react';
import { useAuth } from '../../auth/useAuth';
import { useTranslation } from '../../i18n/useTranslation';
import './domain-components.css';
import { config } from '../../config';

const API = config.leanschoolUrl;

export default function GradeView({ id }) {
  const { t } = useTranslation();
  const { authFetch } = useAuth();
  const [grade, setGrade] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    if (!id) return;
    setLoading(true);
    setError(null);
    authFetch(`${API}/grades/${id}`)
      .then(res => {
        if (!res.ok) throw new Error(`Failed to load grade (${res.status})`);
        return res.json();
      })
      .then(data => setGrade(data))
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
  if (!grade) return null;

  return (
    <div className="dv-card">
      <div className="dv-row">
        <span className="dv-label">{t.domain.common.id}</span>
        <span className="dv-value">{grade.id}</span>
      </div>
      <div className="dv-row">
        <span className="dv-label">{t.domain.grade.grade}</span>
        <span className="dv-value">{grade.grade}</span>
      </div>
      {grade.exam && (
        <div className="dv-row">
          <span className="dv-label">{t.domain.grade.exam}</span>
          <span className="dv-value">{grade.exam.id}</span>
        </div>
      )}
      {grade.student && (
        <div className="dv-row">
          <span className="dv-label">{t.domain.grade.student}</span>
          <span className="dv-value">{grade.student.prename} {grade.student.name}</span>
        </div>
      )}
      <div className="dv-row">
        <span className="dv-label">{t.domain.common.version}</span>
        <span className="dv-value">{grade.version}</span>
      </div>
    </div>
  );
}
