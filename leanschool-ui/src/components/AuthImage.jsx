import { useState, useEffect } from 'react'
import { useAuth } from '../auth/useAuth'

/**
 * Fetches an image through authFetch (adds Bearer token) and renders it
 * via a blob object URL. Handles cleanup on unmount or src change.
 */
export default function AuthImage({ src, alt, className }) {
  const { authFetch } = useAuth()
  const [objectUrl, setObjectUrl] = useState(null)

  useEffect(() => {
    if (!src) return
    let revoke = null
    authFetch(src)
      .then(res => (res.ok ? res.blob() : null))
      .then(blob => {
        if (!blob) return
        revoke = URL.createObjectURL(blob)
        setObjectUrl(revoke)
      })
      .catch(() => {})
    return () => {
      if (revoke) URL.revokeObjectURL(revoke)
      setObjectUrl(null)
    }
  }, [src]) // eslint-disable-line react-hooks/exhaustive-deps

  if (!objectUrl) return null
  return <img src={objectUrl} alt={alt} className={className} />
}
