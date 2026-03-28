import { useState, useEffect } from 'react';
import { useAuth } from '../../auth/useAuth';
import { useTranslation } from '../../i18n/useTranslation';
import './domain-components.css';

const API = import.meta.env.VITE_LEANSCHOOL_URL || 'http://localhost:8080';

export default function ExamView({ id }) {
  const { t } = useTranslation();
  const { authFetch } = useAuth();
  const [exam, setExam] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    if (!id) return;
    setLoading(true);
    setError(null);
    authFetch(`${API}/exams/${id}`)
      .then(res => {
        if (!res.ok) throw new Error(`Failed to load exam (${res.status})`);
        return res.json();
      })
      .then(data => setExam(data))
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
  if (!exam) return null;

  return (
    <div className="dv-card">
      <div className="dv-row">
        <span className="dv-label">{t.domain.common.id}</span>
        <span className="dv-value">{exam.id}</span>
      </div>
      {exam.schoolClass && (
        <div className="dv-row">
          <span className="dv-label">{t.domain.exam.schoolClass}</span>
          <span className="dv-value">{exam.schoolClass.name ?? exam.schoolClass.id}</span>
        </div>
      )}
      <div className="dv-row">
        <span className="dv-label">{t.domain.common.version}</span>
        <span className="dv-value">{exam.version}</span>
      </div>
    </div>
  );
}
