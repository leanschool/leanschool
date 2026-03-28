import { createContext, useContext, useState } from 'react'
import translations from './translations'

const I18nContext = createContext(null)

export function I18nProvider({ children }) {
  const browserLang = navigator.language?.slice(0, 2)
  const initial = translations[browserLang] ? browserLang : 'en'
  const [lang, setLang] = useState(initial)

  return (
    <I18nContext.Provider value={{ lang, setLang, t: translations[lang] }}>
      {children}
    </I18nContext.Provider>
  )
}

export function useTranslation() {
  return useContext(I18nContext)
}
