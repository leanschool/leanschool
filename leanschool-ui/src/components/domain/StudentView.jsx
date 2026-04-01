import { useState, useEffect } from 'react';
import { useAuth } from '../../auth/useAuth';
import { useTranslation } from '../../i18n/useTranslation';
import './domain-components.css';
import { config } from '../../config';

const API = config.leanschoolUrl;

export default function StudentView({ id }) {
  const { t } = useTranslation();
  const { authFetch } = useAuth();
  const [student, setStudent] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    if (!id) return;
    setLoading(true);
    setError(null);
    authFetch(`${API}/students/${id}`)
      .then(res => {
        if (!res.ok) throw new Error(`Failed to load student (${res.status})`);
        return res.json();
      })
      .then(data => setStudent(data))
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
  if (!student) return null;

  return (
    <div className="dv-card">
      <div className="dv-row">
        <span className="dv-label">{t.domain.common.id}</span>
        <span className="dv-value">{student.id}</span>
      </div>
      <div className="dv-row">
        <span className="dv-label">{t.domain.person.name}</span>
        <span className="dv-value">{student.prename} {student.name}</span>
      </div>
      {student.dateOfBirth && (
        <div className="dv-row">
          <span className="dv-label">{t.domain.person.dateOfBirth}</span>
          <span className="dv-value">{student.dateOfBirth.slice(0, 10)}</span>
        </div>
      )}
      {student.address && (
        <div className="dv-row">
          <span className="dv-label">{t.domain.person.address}</span>
          <span className="dv-value">{student.address.name || student.address.id}</span>
        </div>
      )}
      <div className="dv-row">
        <span className="dv-label">{t.domain.common.version}</span>
        <span className="dv-value">{student.version}</span>
      </div>

      {student.currentSchoolClass && (
        <>
          <div className="dv-section-title">{t.domain.student.currentSchoolClass}</div>
          <div className="dv-row">
            <span className="dv-label">{t.domain.student.currentSchoolClass}</span>
            <span className="dv-value">{student.currentSchoolClass.name}</span>
          </div>
        </>
      )}

      {student.guardians && student.guardians.length > 0 && (
        <>
          <div className="dv-section-title">{t.domain.student.guardians}</div>
          <div className="dv-badge-list">
            {student.guardians.map(g => (
              <span key={g.id} className="dv-badge">{g.prename} {g.name}</span>
            ))}
          </div>
        </>
      )}

      {student.pastSchoolClasses && student.pastSchoolClasses.length > 0 && (
        <>
          <div className="dv-section-title">{t.domain.student.pastSchoolClasses}</div>
          <div className="dv-badge-list">
            {student.pastSchoolClasses.map(sc => (
              <span key={sc.id} className="dv-badge">{sc.name}</span>
            ))}
          </div>
        </>
      )}

      {student.grades && student.grades.length > 0 && (
        <>
          <div className="dv-section-title">{t.domain.student.grades}</div>
          <div className="dv-badge-list">
            {student.grades.map((g, i) => (
              <span key={i} className="dv-badge">{g.subject}: {g.value}</span>
            ))}
          </div>
        </>
      )}
    </div>
  );
}
