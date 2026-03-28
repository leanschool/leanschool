import { useState, useEffect } from 'react';
import { useAuth } from '../../auth/useAuth';
import { useTranslation } from '../../i18n/useTranslation';
import './domain-components.css';

const API = import.meta.env.VITE_LEANSCHOOL_URL || 'http://localhost:8080';

export default function TeacherView({ id }) {
  const { t } = useTranslation();
  const { authFetch } = useAuth();
  const [teacher, setTeacher] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    if (!id) return;
    setLoading(true);
    setError(null);
    authFetch(`${API}/teachers/${id}`)
      .then(res => {
        if (!res.ok) throw new Error(`Failed to load teacher (${res.status})`);
        return res.json();
      })
      .then(data => setTeacher(data))
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
  if (!teacher) return null;

  const dob = teacher.dateOfBirth
    ? new Date(teacher.dateOfBirth).toLocaleDateString()
    : '—';
  const since = teacher.atSchoolSince
    ? new Date(teacher.atSchoolSince).toLocaleDateString()
    : '—';
  const until = teacher.atSchoolUntil
    ? new Date(teacher.atSchoolUntil).toLocaleDateString()
    : '—';

  return (
    <div className="dv-card">
      <div className="dv-row">
        <span className="dv-label">{t.domain.common.id}</span>
        <span className="dv-value">{teacher.id}</span>
      </div>
      <div className="dv-row">
        <span className="dv-label">{t.domain.person.prename}</span>
        <span className="dv-value">{teacher.prename}</span>
      </div>
      <div className="dv-row">
        <span className="dv-label">{t.domain.person.name}</span>
        <span className="dv-value">{teacher.name}</span>
      </div>
      <div className="dv-row">
        <span className="dv-label">{t.domain.person.dateOfBirth}</span>
        <span className="dv-value">{dob}</span>
      </div>
      {teacher.sub && (
        <div className="dv-row">
          <span className="dv-label">{t.domain.person.sub}</span>
          <span className="dv-value">{teacher.sub}</span>
        </div>
      )}
      {teacher.iban && (
        <div className="dv-row">
          <span className="dv-label">{t.domain.teacher.iban}</span>
          <span className="dv-value">{teacher.iban}</span>
        </div>
      )}
      <div className="dv-row">
        <span className="dv-label">{t.domain.teacher.atSchoolSince}</span>
        <span className="dv-value">{since}</span>
      </div>
      <div className="dv-row">
        <span className="dv-label">{t.domain.teacher.atSchoolUntil}</span>
        <span className="dv-value">{until}</span>
      </div>
      {teacher.address && (
        <>
          <div className="dv-section-title">{t.domain.person.address}</div>
          <div className="dv-row">
            <span className="dv-label">{t.domain.address.street}</span>
            <span className="dv-value">
              {teacher.address.street} {teacher.address.number}
            </span>
          </div>
          <div className="dv-row">
            <span className="dv-label">{t.domain.address.postalCode}</span>
            <span className="dv-value">{teacher.address.postalCode}</span>
          </div>
        </>
      )}
      <div className="dv-row">
        <span className="dv-label">{t.domain.common.version}</span>
        <span className="dv-value">{teacher.version}</span>
      </div>
    </div>
  );
}
