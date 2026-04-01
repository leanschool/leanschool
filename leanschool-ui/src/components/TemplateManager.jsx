import { useState, useEffect } from 'react'
import { useTranslation } from '../i18n/useTranslation'
import { useAuth } from '../auth/useAuth'
import { config } from '../config'
import './TemplateManager.css'

const EXTRACTION_SERVICE_URL = config.extractionServiceUrl

const EMPTY_FORM = { 
  name: '', 
  templateType: 'text',
  template: '',
  templateVariables: ''
}

export default function TemplateManager({ onBack }) {
  const { t } = useTranslation()
  const { authFetch } = useAuth()
  const [templates, setTemplates] = useState([])
  const [loadState, setLoadState] = useState('loading')
  const [creating, setCreating] = useState(false)
  const [createForm, setCreateForm] = useState(EMPTY_FORM)
  const [createError, setCreateError] = useState('')
  const [editId, setEditId] = useState(null)
  const [editForm, setEditForm] = useState(EMPTY_FORM)
  const [editError, setEditError] = useState('')
  const [deleteError, setDeleteError] = useState('')

  const load = () => {
    setLoadState('loading')
    authFetch(`${EXTRACTION_SERVICE_URL}/templates`)
      .then(r => r.ok ? r.json() : [])
      .then(data => { 
        setTemplates(data ?? []); 
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
      const variables = createForm.templateVariables.split(',').map(v => v.trim()).filter(v => v)
      const res = await authFetch(`${EXTRACTION_SERVICE_URL}/templates`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          name: createForm.name,
          templateType: createForm.templateType,
          template: createForm.template,
          templateVariables: variables,
        }),
      })
      if (!res.ok) throw new Error(res.status)
      setCreateForm(EMPTY_FORM)
      setCreating(false)
      load()
    } catch {
      setCreateError(t.templates.createError)
    }
  }

  // ── edit ──────────────────────────────────────────────────────────────────
  const startEdit = template => {
    setEditId(template.id)
    setEditForm({
      name: template.name,
      templateType: template.templateType,
      template: template.template,
      templateVariables: template.templateVariables.join(', '),
    })
    setEditError('')
  }

  const submitEdit = async e => {
    e.preventDefault()
    setEditError('')
    try {
      const variables = editForm.templateVariables.split(',').map(v => v.trim()).filter(v => v)
      const res = await authFetch(`${EXTRACTION_SERVICE_URL}/templates`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          id: editId,
          name: editForm.name,
          templateType: editForm.templateType,
          template: editForm.template,
          templateVariables: variables,
        }),
      })
      if (!res.ok) throw new Error(res.status)
      setEditId(null)
      load()
    } catch {
      setEditError(t.templates.updateError)
    }
  }

  // ── delete ────────────────────────────────────────────────────────────────
  const deleteTemplate = async id => {
    setDeleteError('')
    try {
      const res = await authFetch(`${EXTRACTION_SERVICE_URL}/templates/${id}`, { method: 'DELETE' })
      if (!res.ok) throw new Error(res.status)
      load()
    } catch {
      setDeleteError(t.templates.deleteError)
    }
  }

  return (
    <div className="tm-page">
      <div className="orb orb-1" />
      <div className="orb orb-2" />

      <div className="tm-header">
        <h2 className="tm-title">{t.templates.title}</h2>
      </div>

      <button className="page-back-btn" onClick={onBack}>← {t.templates.back}</button>

      <div className="tm-content">

        {creating && (
          <form className="tm-card tm-form" onSubmit={submitCreate}>
            <TemplateFormFields form={createForm} onChange={setCreateForm} t={t} autoFocus />
            {createError && <div className="tm-error">{createError}</div>}
            <div className="tm-form-actions">
              <button type="submit" className="cta-button tm-btn">{t.templates.create}</button>
              <button type="button" className="ghost-button tm-btn"
                onClick={() => { setCreating(false); setCreateForm(EMPTY_FORM) }}>
                {t.templates.cancel}
              </button>
            </div>
          </form>
        )}

        {loadState === 'loading' && (
          <div className="tm-center"><div className="spinner" /><span className="tm-hint">{t.templates.loading}</span></div>
        )}
        {loadState === 'error' && (
          <div className="tm-center tm-error">{t.templates.error}</div>
        )}
        {loadState === 'ready' && templates.length === 0 && !creating && (
          <div className="tm-center tm-hint">{t.templates.empty}</div>
        )}

        {deleteError && <div className="tm-error tm-delete-error">{deleteError}</div>}

        {loadState === 'ready' && templates.map(template => (
          <div key={template.id} className="tm-card">
            {editId === template.id ? (
              <form className="tm-form" onSubmit={submitEdit}>
                <TemplateFormFields form={editForm} onChange={setEditForm} t={t} autoFocus />
                {editError && <div className="tm-error">{editError}</div>}
                <div className="tm-form-actions">
                  <button type="submit" className="cta-button tm-btn">{t.templates.save}</button>
                  <button type="button" className="ghost-button tm-btn" onClick={() => setEditId(null)}>
                    {t.templates.cancel}
                  </button>
                </div>
              </form>
            ) : (
              <div className="tm-template-row">
                <div className="tm-template-info">
                  <span className="tm-template-name">{template.name}</span>
                  <span className="tm-template-type-badge">{template.templateType}</span>
                </div>
                <div className="tm-template-right">
                  <span className="tm-template-variables">
                    {template.templateVariables.join(', ')}
                  </span>
                  <button className="tm-icon-btn" onClick={() => startEdit(template)} title={t.templates.edit}>✎</button>
                  <button className="tm-icon-btn tm-icon-btn--danger" onClick={() => deleteTemplate(template.id)} title={t.templates.delete}>✕</button>
                </div>
              </div>
            )}
          </div>
        ))}

        {!creating && (
          <button className="tm-new-btn tm-new-btn--bottom" onClick={() => { setCreating(true); setCreateError('') }}>
            + {t.templates.newTemplate}
          </button>
        )}
      </div>
    </div>
  )
}

function TemplateFormFields({ form, onChange, t, autoFocus }) {
  const set = (key, val) => onChange(f => ({ ...f, [key]: val }))
  return (
    <div className="tm-form-grid">
      <input
        className="field-input tm-field-name"
        placeholder={t.templates.name}
        value={form.name}
        onChange={e => set('name', e.target.value)}
        required
        autoFocus={autoFocus}
      />
      <select
        className="field-input field-select tm-field-type"
        value={form.templateType}
        onChange={e => set('templateType', e.target.value)}
      >
        <option value="text">Text</option>
        <option value="excel">Excel</option>
        <option value="csv">CSV</option>
      </select>
      <textarea
        className="field-input tm-field-template"
        placeholder={t.templates.template}
        value={form.template}
        onChange={e => set('template', e.target.value)}
        required
        rows={4}
      />
      <input
        className="field-input tm-field-variables"
        placeholder={t.templates.variables}
        value={form.templateVariables}
        onChange={e => set('templateVariables', e.target.value)}
      />
    </div>
  )
}