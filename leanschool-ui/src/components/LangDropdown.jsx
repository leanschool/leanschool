import { useState, useEffect, useRef } from 'react'
import { useTranslation } from '../i18n/useTranslation'
import './LangDropdown.css'

export default function LangDropdown({ solo = false }) {
  const { t, lang, setLang } = useTranslation()
  const [open, setOpen] = useState(false)
  const ref = useRef(null)

  useEffect(() => {
    if (!open) return
    const handler = e => { if (!ref.current?.contains(e.target)) setOpen(false) }
    document.addEventListener('mousedown', handler)
    return () => document.removeEventListener('mousedown', handler)
  }, [open])

  return (
    <div className={`lang-dropdown${solo ? ' lang-dropdown--solo' : ''}`} ref={ref}>
      <button className="lang-dropdown-btn" onClick={() => setOpen(o => !o)}>
        {lang.toUpperCase()}
        <span className="lang-dropdown-caret">{open ? '▲' : '▼'}</span>
      </button>
      {open && (
        <div className="lang-dropdown-menu">
          {['en', 'de', 'fr', 'it'].map(l => (
            <button
              key={l}
              className={`lang-dropdown-item${lang === l ? ' active' : ''}`}
              onClick={() => { setLang(l); setOpen(false) }}
            >
              {t.lang[l]}
            </button>
          ))}
        </div>
      )}
    </div>
  )
}
