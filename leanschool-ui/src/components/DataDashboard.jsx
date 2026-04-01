import { useState, useEffect, useCallback } from 'react'
import { useAuth } from '../auth/useAuth'
import { useTranslation } from '../i18n/useTranslation'
import { canReadModel, canWriteModel, canCreateModel } from '../auth/permissions'

import LocationView from './domain/LocationView'
import LocationForm from './domain/LocationForm'
import BuildingView from './domain/BuildingView'
import BuildingForm from './domain/BuildingForm'
import RoomView from './domain/RoomView'
import RoomForm from './domain/RoomForm'
import PostalCodeView from './domain/PostalCodeView'
import PostalCodeForm from './domain/PostalCodeForm'
import CityView from './domain/CityView'
import CityForm from './domain/CityForm'
import AddressView from './domain/AddressView'
import AddressForm from './domain/AddressForm'
import PersonView from './domain/PersonView'
import PersonForm from './domain/PersonForm'
import GuardianView from './domain/GuardianView'
import GuardianForm from './domain/GuardianForm'
import TeacherView from './domain/TeacherView'
import TeacherForm from './domain/TeacherForm'
import StudentView from './domain/StudentView'
import StudentForm from './domain/StudentForm'
import SchoolYearView from './domain/SchoolYearView'
import SchoolYearForm from './domain/SchoolYearForm'
import SchoolClassView from './domain/SchoolClassView'
import SchoolClassForm from './domain/SchoolClassForm'
import CurriculumView from './domain/CurriculumView'
import CurriculumForm from './domain/CurriculumForm'
import SubjectView from './domain/SubjectView'
import SubjectForm from './domain/SubjectForm'
import LessonView from './domain/LessonView'
import LessonForm from './domain/LessonForm'
import ExamView from './domain/ExamView'
import ExamForm from './domain/ExamForm'
import AccountView from './domain/AccountView'
import AccountForm from './domain/AccountForm'
import ReceiptPanel from './domain/ReceiptPanel'

import './DataDashboard.css'

import { config } from '../config'

const API = config.leanschoolUrl

const MODELS = {
  location:    { enabled: false, endpoint: '/locations',      idField: 'id',     idProp: 'id',     label: e => `${e.lon ?? ''}, ${e.lat ?? ''}`,                                            View: LocationView,   Form: LocationForm   },
  building:    { enabled: false, endpoint: '/buildings',      idField: 'id',     idProp: 'id',     label: e => e.name ?? e.id,                                                              View: BuildingView,   Form: BuildingForm   },
  room:        { enabled: false, endpoint: '/rooms',          idField: 'id',     idProp: 'id',     label: e => e.name ?? e.id,                                                              View: RoomView,       Form: RoomForm       },
  postalcode:  { enabled: false, endpoint: '/postal-codes',   idField: 'number', idProp: 'number', label: e => `${e.number} — ${e.city ?? ''}`,                                            View: PostalCodeView, Form: PostalCodeForm },
  city:        { enabled: false, endpoint: '/cities',         idField: 'id',     idProp: 'id',     label: e => e.name ?? e.id,                                                              View: CityView,       Form: CityForm       },
  address:     { enabled: false, endpoint: '/addresses',      idField: 'id',     idProp: 'id',     label: e => [e.street, e.number].filter(Boolean).join(' ') || e.id,                     View: AddressView,    Form: AddressForm    },
  person:      { enabled: true, endpoint: '/persons',        idField: 'id',     idProp: 'id',     label: e => [e.prename, e.name].filter(Boolean).join(' ') || e.id,                      View: PersonView,     Form: PersonForm     },
  guardian:    { enabled: false, endpoint: '/guardians',      idField: 'id',     idProp: 'id',     label: e => [e.prename, e.name].filter(Boolean).join(' ') || e.id,                      View: GuardianView,   Form: GuardianForm   },
  teacher:     { enabled: true, endpoint: '/teachers',       idField: 'id',     idProp: 'id',     label: e => [e.prename, e.name].filter(Boolean).join(' ') || e.id,                      View: TeacherView,    Form: TeacherForm    },
  student:     { enabled: false, endpoint: '/students',       idField: 'id',     idProp: 'id',     label: e => [e.prename, e.name].filter(Boolean).join(' ') || e.id,                      View: StudentView,    Form: StudentForm    },
  schoolyear:  { enabled: false, endpoint: '/school-years',   idField: 'id',     idProp: 'id',     label: e => `${e.from ?? ''} – ${e.to ?? ''}`,                                          View: SchoolYearView, Form: SchoolYearForm },
  schoolclass: { enabled: true, endpoint: '/school-classes', idField: 'id',     idProp: 'id',     label: e => e.shortcut ? `${e.name ?? ''} (${e.shortcut})` : (e.name ?? e.id),         View: SchoolClassView,Form: SchoolClassForm},
  curriculum:  { enabled: false, endpoint: '/curricula',      idField: 'id',     idProp: 'id',     label: e => e.name ?? e.id,                                                              View: CurriculumView, Form: CurriculumForm },
  subject:     { enabled: false, endpoint: '/subjects',       idField: 'id',     idProp: 'id',     label: e => e.name ?? e.id,                                                              View: SubjectView,    Form: SubjectForm    },
  lesson:      { enabled: false, endpoint: '/lessons',        idField: 'id',     idProp: 'id',     label: e => e.id,                                                                        View: LessonView,     Form: LessonForm     },
  exam:        { enabled: false, endpoint: '/exams',          idField: 'id',     idProp: 'id',     label: e => e.id,                                                                        View: ExamView,       Form: ExamForm       },
  grade:       { enabled: false, endpoint: '/grades',         idField: 'id',     idProp: 'id',     label: e => e.grade != null ? String(e.grade) : e.id,                                    View: ExamView,       Form: ExamForm       },
  account:     { enabled: true,  endpoint: '/accounts',       idField: 'id',     idProp: 'id',     label: e => e.name != null ? String(e.name) : e.id,                                      View: AccountView,    Form: AccountForm    },
  receipt:     { enabled: true,  endpoint: '/receipts',       idField: 'id',     idProp: 'id',     label: e => e.id,                                                                         Custom: ReceiptPanel },
}

const MODEL_GROUPS = [
  { enabled: false, key: 'infrastructure', models: ['location', 'building', 'room'].filter(k => MODELS[k].enabled) },
  { enabled: false, key: 'geography',      models: ['postalcode', 'city', 'address'].filter(k => MODELS[k].enabled) },
  { enabled: true, key: 'people',         models: ['person', 'guardian', 'teacher', 'student'].filter(k => MODELS[k].enabled) },
  { enabled: true, key: 'school',         models: ['schoolyear', 'schoolclass', 'curriculum', 'subject', 'lesson', 'exam', 'grade'].filter(k => MODELS[k].enabled) },
  { enabled: true, key: 'finances',       models: ['account', 'receipt'].filter(k => MODELS[k].enabled) },
]

export default function DataDashboard({ onBack }) {
  const { t } = useTranslation()
  const { user, authFetch } = useAuth()

  const [selectedModel, setSelectedModel] = useState(null)
  const [mode, setMode] = useState('list') // list | view | create | edit
  const [editId, setEditId] = useState(null)
  const [entities, setEntities] = useState([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(null)
  const [deleteTarget, setDeleteTarget] = useState(null)
  const [deleting, setDeleting] = useState(false)

  const accessibleGroups = MODEL_GROUPS.map(g => ({
    ...g,
    models: g.models.filter(k => canReadModel(user, k)),
  })).filter(g => g.models.length > 0).filter(g => g.enabled)

  // Auto-select the first accessible model on mount.
  useEffect(() => {
    const first = accessibleGroups.flatMap(g => g.models)[0]
    if (first) setSelectedModel(first)
  }, []) // eslint-disable-line react-hooks/exhaustive-deps

  const fetchList = useCallback(() => {
    if (!selectedModel) return
    const config = MODELS[selectedModel]
    setLoading(true)
    setError(null)
    authFetch(`${API}${config.endpoint}`)
      .then(r => r.ok ? r.json() : [])
      .then(data => setEntities(Array.isArray(data) ? data : []))
      .catch(() => setError(t.dataDashboard.loadError))
      .finally(() => setLoading(false))
  }, [selectedModel]) // eslint-disable-line react-hooks/exhaustive-deps

  useEffect(() => {
    if (mode === 'list') fetchList()
  }, [selectedModel, mode]) // eslint-disable-line react-hooks/exhaustive-deps

  function selectModel(key) {
    setSelectedModel(key)
    setMode('list')
    setEditId(null)
  }

  function goList() {
    setMode('list')
    setEditId(null)
  }

  async function handleDelete() {
    if (!deleteTarget || !selectedModel) return
    const config = MODELS[selectedModel]
    setDeleting(true)
    try {
      await authFetch(`${API}${config.endpoint}/${deleteTarget}`, { method: 'DELETE' })
      setDeleteTarget(null)
      fetchList()
    } catch {
      // stay open, user can retry
    } finally {
      setDeleting(false)
    }
  }

  // ── Sidebar ───────────────────────────────────────────────────────────────

  const sidebar = (
    <aside className="dd-sidebar">
      <div className="dd-sidebar-heading">{t.dataDashboard.title}</div>
      {accessibleGroups.map(group => (
        <div key={group.key} className="dd-group">
          <div className="dd-group-label">{t.dataDashboard.groups[group.key]}</div>
          {group.models.map(key => (
            <button
              key={key}
              className={`dd-model-btn${selectedModel === key ? ' dd-model-btn--active' : ''}`}
              onClick={() => selectModel(key)}
            >
              {t.dataDashboard.models[key]}
            </button>
          ))}
        </div>
      ))}
    </aside>
  )

  // ── Main content ──────────────────────────────────────────────────────────

  let content

  if (!selectedModel) {
    content = <p className="dd-empty">{t.dataDashboard.selectModel}</p>
  } else {
    const config = MODELS[selectedModel]
    const { View, Form, Custom } = config
    const idProp = config.idProp

    if (Custom) {
      content = <Custom />
    } else if (mode === 'view') {
      content = (
        <div className="dd-detail">
          <button className="ghost-button dd-back-btn" onClick={goList}>
            ← {t.dataDashboard.backToList}
          </button>
          <View {...{ [idProp]: editId }} />
        </div>
      )
    } else if (mode === 'create' || mode === 'edit') {
      const formIdProps = mode === 'edit' ? { [idProp]: editId } : {}
      content = (
        <div className="dd-detail">
          <button className="ghost-button dd-back-btn" onClick={goList}>
            ← {t.dataDashboard.backToList}
          </button>
          <Form
            {...formIdProps}
            persist
            onSave={goList}
            onCancel={goList}
          />
        </div>
      )
    } else {
      // list mode
      const canWrite = canWriteModel(user, selectedModel)
      const canCreate = canCreateModel(user, selectedModel)

      content = (
        <div className="dd-list">
          <div className="dd-list-header">
            <h2 className="dd-list-title">{t.dataDashboard.models[selectedModel]}</h2>
            {canCreate && (
              <button className="cta-button" onClick={() => setMode('create')}>
                + {t.dataDashboard.create}
              </button>
            )}
          </div>

          {loading && (
            <div className="dd-loading"><div className="spinner" /></div>
          )}
          {error && <div className="dd-error">{error}</div>}

          {!loading && !error && (
            entities.length === 0
              ? <p className="dd-empty">{t.dataDashboard.empty}</p>
              : (
                <div className="dd-entities">
                  {entities.map(entity => {
                    const id = entity[config.idField]
                    return (
                      <div key={id} className="dd-entity-row">
                        <div className="dd-entity-info">
                          <span className="dd-entity-id">{String(id)}</span>
                          <span className="dd-entity-desc">{config.label(entity)}</span>
                        </div>
                        <div className="dd-entity-actions">
                          <button
                            className="ghost-button dd-act-btn"
                            onClick={() => { setEditId(id); setMode('view') }}
                          >
                            {t.dataDashboard.view}
                          </button>
                          {canWrite && (
                            <>
                              <button
                                className="ghost-button dd-act-btn"
                                onClick={() => { setEditId(id); setMode('edit') }}
                              >
                                {t.dataDashboard.edit}
                              </button>
                              <button
                                className="ghost-button dd-act-btn dd-act-danger"
                                onClick={() => setDeleteTarget(id)}
                              >
                                {t.dataDashboard.delete}
                              </button>
                            </>
                          )}
                        </div>
                      </div>
                    )
                  })}
                </div>
              )
          )}

          {deleteTarget !== null && (
            <div className="dd-overlay">
              <div className="dd-confirm">
                <p className="dd-confirm-msg">{t.dataDashboard.deleteConfirm}</p>
                <div className="dd-confirm-actions">
                  <button
                    className="cta-button dd-btn-danger"
                    onClick={handleDelete}
                    disabled={deleting}
                  >
                    {deleting ? '…' : t.dataDashboard.confirmDelete}
                  </button>
                  <button
                    className="ghost-button"
                    onClick={() => setDeleteTarget(null)}
                    disabled={deleting}
                  >
                    {t.dataDashboard.cancel}
                  </button>
                </div>
              </div>
            </div>
          )}
        </div>
      )
    }
  }

  return (
    <div className="dd-page">
      <div className="orb orb-1" />
      <div className="orb orb-2" />

      <nav className="dd-nav">
        <img src="/logo.svg" alt="leanschool" className="dd-nav-logo" />
        <button className="ghost-button dd-nav-back" onClick={onBack}>
          ← {t.dataDashboard.backToDashboard}
        </button>
      </nav>

      <div className="dd-layout">
        {sidebar}
        <main className="dd-content">{content}</main>
      </div>
    </div>
  )
}
