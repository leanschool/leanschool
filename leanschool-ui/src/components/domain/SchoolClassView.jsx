import { useState, useEffect } from 'react';
import { useAuth } from '../../auth/useAuth';
import { useTranslation } from '../../i18n/useTranslation';
import './domain-components.css';
import { config } from '../../config';

const API = config.leanschoolUrl;

export default function SchoolClassView({ id }) {
  const { t } = useTranslation();
  const { authFetch } = useAuth();
  const [schoolClass, setSchoolClass] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    if (!id) return;
    setLoading(true);
    setError(null);
    authFetch(`${API}/school-classes/${id}`)
      .then(res => {
        if (!res.ok) throw new Error(`Failed to load school class (${res.status})`);
        return res.json();
      })
      .then(data => setSchoolClass(data))
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
  if (!schoolClass) return null;

  return (
    <div className="dv-card">
      <div className="dv-row">
        <span className="dv-label">{t.domain.common.id}</span>
        <span className="dv-value">{schoolClass.id}</span>
      </div>
      <div className="dv-row">
        <span className="dv-label">{t.domain.schoolClass.name}</span>
        <span className="dv-value">{schoolClass.name}</span>
      </div>
      {schoolClass.shortcut && (
        <div className="dv-row">
          <span className="dv-label">{t.domain.schoolClass.shortcut}</span>
          <span className="dv-value">{schoolClass.shortcut}</span>
        </div>
      )}
      {schoolClass.classroom && (
        <div className="dv-row">
          <span className="dv-label">{t.domain.schoolClass.classroom}</span>
          <span className="dv-value">{schoolClass.classroom.name || schoolClass.classroom.id}</span>
        </div>
      )}
      {schoolClass.schoolYear && (
        <div className="dv-row">
          <span className="dv-label">{t.domain.schoolClass.schoolYear}</span>
          <span className="dv-value">
            {schoolClass.schoolYear.from
              ? `${schoolClass.schoolYear.from.slice(0, 10)} – ${schoolClass.schoolYear.to.slice(0, 10)}`
              : schoolClass.schoolYear.id}
          </span>
        </div>
      )}
      <div className="dv-row">
        <span className="dv-label">{t.domain.common.version}</span>
        <span className="dv-value">{schoolClass.version}</span>
      </div>

      {schoolClass.teachers && schoolClass.teachers.length > 0 && (
        <>
          <div className="dv-section-title">{t.domain.schoolClass.teachers}</div>
          <div className="dv-badge-list">
            {schoolClass.teachers.map(t => (
              <span key={t.id} className="dv-badge">{t.prename} {t.name}</span>
            ))}
          </div>
        </>
      )}

      {schoolClass.students && schoolClass.students.length > 0 && (
        <>
          <div className="dv-section-title">{t.domain.schoolClass.students}</div>
          <div className="dv-badge-list">
            {schoolClass.students.map(s => (
              <span key={s.id} className="dv-badge">{s.prename} {s.name}</span>
            ))}
          </div>
        </>
      )}
    </div>
  );
}
