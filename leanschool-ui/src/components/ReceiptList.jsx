import { useState, useEffect } from 'react'
import { useTranslation } from '../i18n/useTranslation'
import { useAuth } from '../auth/useAuth'
import AuthImage from './AuthImage'
import ExtractionTemplateSelector from './domain/ExtractionTemplateSelector'
import './ReceiptList.css'

const LEANSCHOOL_URL   = import.meta.env.VITE_LEANSCHOOL_URL   || 'http://localhost:8080'
const FILE_SERVICE_URL = import.meta.env.VITE_FILE_SERVICE_URL || 'http://localhost:8083'
const EXTRACTION_SERVICE_URL = import.meta.env.VITE_EXTRACTION_SERVICE_URL || 'http://localhost:8084'

const STATUS_ORDER = ['unsubmitted', 'submitted', 'accepted', 'declined']

export default function ReceiptList({ onBack, canSubmit = false, embedded = false }) {
  const { t } = useTranslation()
  const { authFetch } = useAuth()
  const [receipts, setReceipts] = useState([])
  const [loadState, setLoadState] = useState('loading') // loading | ready | error
  const [expanded, setExpanded] = useState(null)
  const [selected, setSelected] = useState(new Set())
  const [activeTab, setActiveTab] = useState(null)
  const [exporting, setExporting] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [showTemplateSelector, setShowTemplateSelector] = useState(false)
  const [exportData, setExportData] = useState([])

  const load = () => {
    setLoadState('loading')
    authFetch(`${LEANSCHOOL_URL}/receipts`)
      .then(r => { if (!r.ok) throw new Error(r.status); return r.json() })
      .then(data => {
        setReceipts(data)
        setLoadState('ready')
        // Pick first populated tab in lifecycle order
        const firstTab = STATUS_ORDER.find(s => data.some(r => r.status === s))
        setActiveTab(prev => prev ?? firstTab ?? STATUS_ORDER[0])
      })
      .catch(() => setLoadState('error'))
  }

  useEffect(load, [])

  // Tabs that have at least one receipt
  const tabs = STATUS_ORDER.filter(s => receipts.some(r => r.status === s))
  const visible = receipts.filter(r => r.status === activeTab)

  const switchTab = tab => {
    setActiveTab(tab)
    setSelected(new Set())
    setExpanded(null)
  }

  const toggle = id => setExpanded(prev => prev === id ? null : id)

  const toggleSelect = (e, id) => {
    e.stopPropagation()
    setSelected(prev => {
      const next = new Set(prev)
      next.has(id) ? next.delete(id) : next.add(id)
      return next
    })
  }

  const fmt = iso => new Date(iso).toLocaleString(undefined, {
    year: 'numeric', month: 'short', day: 'numeric',
    hour: '2-digit', minute: '2-digit',
  })

  const submitReceipts = async () => {
    setSubmitting(true)
    try {
      const res = await authFetch(`${LEANSCHOOL_URL}/receipts/submit`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ ids: [...selected] }),
      })
      if (!res.ok) throw new Error(res.status)
      setSelected(new Set())
      load()
    } catch {
      // silently ignore
    } finally {
      setSubmitting(false)
    }
  }

  const exportExcel = async () => {
    if (selected.size === 0) return;
    
    setExporting(true)
    try {
      // Fetch receipt details for selected IDs
      const receiptPromises = Array.from(selected).map(id => 
        authFetch(`${LEANSCHOOL_URL}/receipts/${id}`)
      )
      
      const receipts = await Promise.all(receiptPromises)
      const receiptData = await Promise.all(receipts.map(r => r.json()))
      
      // Prepare data for template
      const exportData = receiptData.map((receipt) => ({
        id: receipt.id,
        date: receipt.date,
        amount: receipt.amount,
        merchant: receipt.merchant,
        category: receipt.category,
        status: receipt.status,
        taxes: receipt.taxes,
        total: receipt.total,
        items: receipt.items
      }))
      
      setExportData(exportData)
      setShowTemplateSelector(true)
      
    } catch (err) {
      console.error('Failed to prepare export data:', err)
    } finally {
      setExporting(false)
    }
  }

  const handleTemplateSelect = (templateId, templateData) => {
    setShowTemplateSelector(false)
  }

  return (
    <div className={embedded ? 'rl-embedded' : 'list-page'}>
      {!embedded && <div className="orb orb-1" />}
      {!embedded && <div className="orb orb-2" />}

      {!embedded && (
        <div className="list-header">
          <h2 className="list-title">{t.list.title}</h2>
        </div>
      )}

      {loadState === 'loading' && (
        <div className="list-center">
          <div className="spinner" />
          <span className="list-hint">{t.list.loading}</span>
        </div>
      )}

      {loadState === 'error' && (
        <div className="list-center list-error">{t.list.error}</div>
      )}

      {loadState === 'ready' && receipts.length === 0 && (
        <div className="list-center list-hint">{t.list.empty}</div>
      )}

      {loadState === 'ready' && receipts.length > 0 && (
        <>
          <div className="list-tabs">
            {tabs.map(s => (
              <button
                key={s}
                className={`list-tab${activeTab === s ? ' list-tab--active' : ''}`}
                onClick={() => switchTab(s)}
              >
                {t.list.status[s]}
                <span className="list-tab-count">
                  {receipts.filter(r => r.status === s).length}
                </span>
              </button>
            ))}
          </div>

          <div className="list-content">
            {visible.map(r => (
              <div
                key={r.id}
                className={`receipt-card${expanded === r.id ? ' receipt-card--open' : ''}${selected.has(r.id) ? ' receipt-card--selected' : ''}`}
                onClick={() => toggle(r.id)}
              >
                <div className="receipt-card-header">
                  <input
                    type="checkbox"
                    className="receipt-card-checkbox"
                    checked={selected.has(r.id)}
                    onChange={e => toggleSelect(e, r.id)}
                    onClick={e => e.stopPropagation()}
                  />
                  <div className="receipt-card-left">
                    <span className="receipt-card-date">{fmt(r.time)}</span>
                    {r.receiptOwnerId && (
                      <span className="receipt-card-owner">◈ {r.receiptOwnerId}</span>
                    )}
                  </div>
                  <div className="receipt-card-right">
                    <span className="receipt-card-total">{r.totalPrice.toFixed(2)}</span>
                    <span className="receipt-card-items">{r.items?.length ?? 0} {t.list.items}</span>
                    <span className="receipt-card-chevron">{expanded === r.id ? '▲' : '▼'}</span>
                  </div>
                </div>

                {expanded === r.id && (r.items?.length > 0 || r.fileId) && (
                  <div className="receipt-card-body">
                    {r.items?.length > 0 && (
                      <>
                        <div className="receipt-items-table">
                          <div className="receipt-items-head">
                            <span>{t.list.itemName}</span>
                            <span>{t.list.itemQty}</span>
                            <span>{t.list.itemPrice}</span>
                          </div>
                          {r.items.map((it, i) => (
                            <div key={i} className="receipt-items-row">
                              <span>{it.name}</span>
                              <span>{it.amount}</span>
                              <span>{it.price.toFixed(2)}</span>
                            </div>
                          ))}
                        </div>
                        {r.taxes > 0 && (
                          <div className="receipt-card-taxes">
                            {t.list.taxes}: {r.taxes.toFixed(2)}
                          </div>
                        )}
                      </>
                    )}
                    {r.fileId && (
                      <div className="receipt-image-wrap">
                        <AuthImage
                          src={`${FILE_SERVICE_URL}/files/${r.fileId}`}
                          alt="receipt"
                          className="receipt-image"
                        />
                      </div>
                    )}
                  </div>
                )}
              </div>
            ))}
          </div>
        </>
      )}
      {selected.size > 0 && (
        <div className="list-bottom-actions">
          {canSubmit && (
            <button className="ghost-button list-action-btn" onClick={submitReceipts} disabled={submitting}>
              {submitting ? t.list.submitting : `${t.list.submit} (${selected.size})`}
            </button>
          )}
          <button className="cta-button list-action-btn" onClick={exportExcel} disabled={exporting}>
            {exporting ? t.list.exporting : `${t.list.exportExcel} (${selected.size})`}
          </button>
        </div>
      )}
      {!embedded && <button className="page-back-btn" onClick={onBack}>← {t.list.back}</button>}
      
      {showTemplateSelector && (
        <div className="template-selector-overlay">
          <div className="template-selector-container">
            <ExtractionTemplateSelector
              onSelect={handleTemplateSelect}
              dataVariables={{
                receipts: JSON.stringify(exportData),
                count: exportData.length,
                totalAmount: exportData.reduce((sum, r) => sum + parseFloat(r.amount || 0), 0),
                currency: 'CHF'
              }}
              persist={true}
            />
            <button 
              className="close-template-selector"
              onClick={() => setShowTemplateSelector(false)}
            >
              ×
            </button>
          </div>
        </div>
      )}
    </div>
  )
}
