import { useState, useEffect } from 'react'
import { useTranslation } from '../i18n/useTranslation'
import { useAuth } from '../auth/useAuth'
import './SchoolClassManager.css'

const API = import.meta.env.VITE_API_URL || 'http://localhost:8080'

const EMPTY_CLASS = { name: '', students: [], teachers: [] }

export default function SchoolClassManager({ onBack }) {
  const { t } = useTranslation()
  const { authFetch } = useAuth()
  const [classes, setClasses] = useState([])
  const [loadState, setLoadState] = useState('loading')
  const [registeredTeachers, setRegisteredTeachers] = useState([])
  const [creating, setCreating] = useState(false)
  const [createForm, setCreateForm] = useState(EMPTY_CLASS)
  const [createError, setCreateError] = useState('')
  const [editId, setEditId] = useState(null)
  const [editForm, setEditForm] = useState(EMPTY_CLASS)
  const [editError, setEditError] = useState('')
  const [deleteError, setDeleteError] = useState('')

  const load = () => {
    setLoadState('loading')
    Promise.all([
      authFetch(`${API}/school-classes`)
        .then(r => { if (!r.ok) throw new Error(r.status); return r.json() }),
      authFetch(`${API}/users/teachers`)
        .then(r => r.ok ? r.json() : [])
        .catch(() => []),
    ])
      .then(([data, teachers]) => {
        setClasses(data ?? [])
        setRegisteredTeachers(teachers ?? [])
        setLoadState('ready')
      })
      .catch(() => setLoadState('error'))
  }

  useEffect(load, []) // eslint-disable-line react-hooks/exhaustive-deps

  // ── create ────────────────────────────────────────────────────────────────
  const submitCreate = async e => {
    e.preventDefault()
    setCreateError('')
    try {
      const res = await authFetch(`${API}/school-classes`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(createForm),
      })
      if (!res.ok) throw new Error(res.status)
      setCreateForm(EMPTY_CLASS)
      setCreating(false)
      load()
    } catch {
      setCreateError(t.schoolClasses.createError)
    }
  }

  // ── edit ──────────────────────────────────────────────────────────────────
  const startEdit = sc => {
    setEditId(sc.id)
    setEditForm({ name: sc.name, students: sc.students ?? [], teachers: sc.teachers ?? [] })
    setEditError('')
  }

  const submitEdit = async e => {
    e.preventDefault()
    setEditError('')
    try {
      const res = await authFetch(`${API}/school-classes/${editId}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(editForm),
      })
      if (!res.ok) throw new Error(res.status)
      setEditId(null)
      load()
    } catch {
      setEditError(t.schoolClasses.updateError)
    }
  }

  // ── delete ────────────────────────────────────────────────────────────────
  const deleteClass = async id => {
    setDeleteError('')
    try {
      const res = await authFetch(`${API}/school-classes/${id}`, { method: 'DELETE' })
      if (!res.ok) throw new Error(res.status)
      load()
    } catch {
      setDeleteError(t.schoolClasses.deleteError)
    }
  }

  return (
    <div className="scm-page">
      <div className="orb orb-1" />
      <div className="orb orb-2" />

      <div className="am-header">
        <h2 className="am-title">{t.schoolClasses.title}</h2>
      </div>

      <button className="page-back-btn" onClick={onBack}>← {t.schoolClasses.back}</button>

      <div className="am-content">

        {creating && (
          <form className="am-card am-form" onSubmit={submitCreate}>
            <ClassForm
              form={createForm}
              onChange={setCreateForm}
              registeredTeachers={registeredTeachers}
              t={t}
              autoFocus
            />
            {createError && <div className="am-error">{createError}</div>}
            <div className="am-form-actions">
              <button type="submit" className="cta-button am-btn">{t.schoolClasses.create}</button>
              <button type="button" className="ghost-button am-btn"
                onClick={() => { setCreating(false); setCreateForm(EMPTY_CLASS) }}>
                {t.schoolClasses.cancel}
              </button>
            </div>
          </form>
        )}

        {loadState === 'loading' && (
          <div className="am-center"><div className="spinner" /><span className="am-hint">{t.schoolClasses.loading}</span></div>
        )}
        {loadState === 'error' && (
          <div className="am-center am-error">{t.schoolClasses.error}</div>
        )}
        {loadState === 'ready' && classes.length === 0 && !creating && (
          <div className="am-center am-hint">{t.schoolClasses.empty}</div>
        )}

        {deleteError && <div className="am-error am-delete-error">{deleteError}</div>}

        {loadState === 'ready' && classes.map(sc => (
          <div key={sc.id} className="am-card">
            {editId === sc.id ? (
              <form className="am-form" onSubmit={submitEdit}>
                <ClassForm
                  form={editForm}
                  onChange={setEditForm}
                  registeredTeachers={registeredTeachers}
                  t={t}
                  autoFocus
                />
                {editError && <div className="am-error">{editError}</div>}
                <div className="am-form-actions">
                  <button type="submit" className="cta-button am-btn">{t.schoolClasses.save}</button>
                  <button type="button" className="ghost-button am-btn" onClick={() => setEditId(null)}>
                    {t.schoolClasses.cancel}
                  </button>
                </div>
              </form>
            ) : (
              <div className="scm-class-row">
                <div className="scm-class-info">
                  <span className="scm-class-name">{sc.name}</span>
                  <div className="scm-persons-row">
                    <PersonBadges label={t.schoolClasses.teachers} persons={sc.teachers ?? []} color="teacher" />
                    <PersonBadges label={t.schoolClasses.students} persons={sc.students ?? []} color="student" />
                  </div>
                </div>
                <div className="am-account-right">
                  <button className="am-icon-btn" onClick={() => startEdit(sc)} title={t.schoolClasses.edit}>✎</button>
                  <button className="am-icon-btn am-icon-btn--danger" onClick={() => deleteClass(sc.id)} title={t.schoolClasses.delete}>✕</button>
                </div>
              </div>
            )}
          </div>
        ))}

        {!creating && (
          <button className="am-new-btn am-new-btn--bottom" onClick={() => { setCreating(true); setCreateError('') }}>
            + {t.schoolClasses.newClass}
          </button>
        )}
      </div>
    </div>
  )
}

// ── ClassForm ────────────────────────────────────────────────────────────────

function ClassForm({ form, onChange, registeredTeachers, t, autoFocus }) {
  // Teachers: select from registered Keycloak users
  const addTeacher = sub => {
    if (!sub) return
    const teacher = registeredTeachers.find(rt => rt.sub === sub)
    if (!teacher) return
    // Avoid duplicates
    if (form.teachers.some(t => t.sub === sub)) return
    onChange(f => ({
      ...f,
      teachers: [...f.teachers, { id: teacher.sub, sub: teacher.sub, name: teacher.name }],
    }))
  }

  const removeTeacher = idx => {
    onChange(f => ({ ...f, teachers: f.teachers.filter((_, i) => i !== idx) }))
  }

  // Students: free-text entry
  const addStudent = () => {
    onChange(f => ({ ...f, students: [...f.students, { id: '', name: '' }] }))
  }

  const updateStudent = (idx, name) => {
    onChange(f => {
      const arr = [...f.students]
      arr[idx] = { ...arr[idx], name }
      return { ...f, students: arr }
    })
  }

  const removeStudent = idx => {
    onChange(f => ({ ...f, students: f.students.filter((_, i) => i !== idx) }))
  }

  const availableTeachers = registeredTeachers.filter(
    rt => !form.teachers.some(t => t.sub === rt.sub)
  )

  return (
    <div className="scm-form-body">
      <input
        className="field-input"
        placeholder={t.schoolClasses.className}
        value={form.name}
        onChange={e => onChange(f => ({ ...f, name: e.target.value }))}
        required
        autoFocus={autoFocus}
      />

      {/* Teachers — select from registered Keycloak users */}
      <div className="scm-person-list">
        <div className="scm-person-list-header">
          <span className="scm-person-group-label">{t.schoolClasses.teachers}</span>
        </div>
        {form.teachers.map((teacher, i) => (
          <div key={teacher.sub || i} className="scm-person-row">
            <span className="scm-person-name">{teacher.name}</span>
            <button type="button" className="am-icon-btn am-icon-btn--danger" onClick={() => removeTeacher(i)}>✕</button>
          </div>
        ))}
        {availableTeachers.length > 0 ? (
          <select
            className="field-input field-select scm-teacher-select"
            value=""
            onChange={e => addTeacher(e.target.value)}
          >
            <option value="">{t.schoolClasses.selectTeacher}</option>
            {availableTeachers.map(rt => (
              <option key={rt.sub} value={rt.sub}>{rt.name}</option>
            ))}
          </select>
        ) : (
          registeredTeachers.length === 0 && (
            <span className="scm-hint">{t.schoolClasses.noTeachers}</span>
          )
        )}
      </div>

      {/* Students — free-text */}
      <PersonList
        label={t.schoolClasses.students}
        persons={form.students}
        onAdd={addStudent}
        onUpdate={(i, name) => updateStudent(i, name)}
        onRemove={removeStudent}
        t={t}
      />
    </div>
  )
}

function PersonList({ label, persons, onAdd, onUpdate, onRemove, t }) {
  return (
    <div className="scm-person-list">
      <div className="scm-person-list-header">
        <span className="scm-person-group-label">{label}</span>
        <button type="button" className="scm-add-person-btn" onClick={onAdd}>
          + {t.schoolClasses.addPerson}
        </button>
      </div>
      {persons.map((p, i) => (
        <div key={i} className="scm-person-row">
          <input
            className="field-input scm-person-input"
            placeholder={t.schoolClasses.personName}
            value={p.name}
            onChange={e => onUpdate(i, e.target.value)}
          />
          <button type="button" className="am-icon-btn am-icon-btn--danger" onClick={() => onRemove(i)}>✕</button>
        </div>
      ))}
    </div>
  )
}

function PersonBadges({ label, persons, color }) {
  if (!persons.length) return null
  return (
    <div className="scm-badge-group">
      <span className="scm-badge-group-label">{label}:</span>
      {persons.map(p => (
        <span key={p.id || p.sub} className={`scm-person-badge scm-person-badge--${color}`}>{p.name}</span>
      ))}
    </div>
  )
}
