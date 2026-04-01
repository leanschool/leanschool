import { useState, useEffect } from 'react';
import { useAuth } from '../../auth/useAuth';
import { useTranslation } from '../../i18n/useTranslation';
import './domain-components.css';
import { config } from '../../config';

const API = config.leanschoolUrl;

export default function LessonView({ id }) {
  const { t } = useTranslation();
  const { authFetch } = useAuth();
  const [lesson, setLesson] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    if (!id) return;
    setLoading(true);
    setError(null);
    authFetch(`${API}/lessons/${id}`)
      .then(res => {
        if (!res.ok) throw new Error(`Failed to load lesson (${res.status})`);
        return res.json();
      })
      .then(data => setLesson(data))
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
  if (!lesson) return null;

  return (
    <div className="dv-card">
      <div className="dv-row">
        <span className="dv-label">{t.domain.common.id}</span>
        <span className="dv-value">{lesson.id}</span>
      </div>
      {lesson.teacher && (
        <div className="dv-row">
          <span className="dv-label">{t.domain.lesson.teacher}</span>
          <span className="dv-value">{lesson.teacher.prename} {lesson.teacher.name}</span>
        </div>
      )}
      {lesson.schoolClass && (
        <div className="dv-row">
          <span className="dv-label">{t.domain.lesson.schoolClass}</span>
          <span className="dv-value">{lesson.schoolClass.name ?? lesson.schoolClass.id}</span>
        </div>
      )}
      {lesson.subject && (
        <div className="dv-row">
          <span className="dv-label">{t.domain.lesson.subject}</span>
          <span className="dv-value">{lesson.subject.name ?? lesson.subject.id}</span>
        </div>
      )}
      {lesson.room && (
        <div className="dv-row">
          <span className="dv-label">{t.domain.lesson.room}</span>
          <span className="dv-value">
            {lesson.room.name}
            {lesson.room.roomType && ` (${t.domain.roomTypes[lesson.room.roomType] || lesson.room.roomType})`}
          </span>
        </div>
      )}
      <div className="dv-row">
        <span className="dv-label">{t.domain.common.version}</span>
        <span className="dv-value">{lesson.version}</span>
      </div>
    </div>
  );
}
