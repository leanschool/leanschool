import { useRef, useState } from 'react'
import { Camera, CameraResultType, CameraSource } from '@capacitor/camera'
import { useTranslation } from '../i18n/useTranslation'
import { useAuth } from '../auth/useAuth'
import { config } from '../config'
import './ScanReceipt.css'

const TESSERACT_URL    = config.receiptReaderUrl
const FILE_SERVICE_URL = config.fileServiceUrl

export default function ScanReceipt({ onExtracted, onCancel }) {
  const { t } = useTranslation()
  const { authFetch } = useAuth()
  const [state, setState] = useState('idle') // idle | processing | error
  const [errorMsg, setErrorMsg] = useState('')
  const tesseractInputRef = useRef(null)

  const sendImage = async (blob, filename, endpointURL) => {
    setState('processing')
    setErrorMsg('')
    try {
      // Step 1: upload the raw image to the file-service and get a stable file ID.
      const fileForm = new FormData()
      fileForm.append('file', blob, filename)
      const uploadRes = await authFetch(`${FILE_SERVICE_URL}/files`, { method: 'POST', body: fileForm })
      if (!uploadRes.ok) throw new Error(`file upload failed: ${uploadRes.status}`)
      const { id: fileId } = await uploadRes.json()

      // Step 2: send only the file ID to the receipt-reader for OCR.
      const res = await authFetch(`${endpointURL}/receipts/extract`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ fileId }),
      })
      if (!res.ok) throw new Error(`${res.status}`)

      const receipt = await res.json()
      onExtracted(receipt)
    } catch (err) {
      console.error(err)
      setState('error')
      setErrorMsg(t.scan.error)
    }
  }

  const handleCamera = async () => {
    try {
      const photo = await Camera.getPhoto({
        resultType: CameraResultType.Base64,
        source: CameraSource.Camera,
        quality: 90,
        allowEditing: false,
      })
      const blob = base64ToBlob(photo.base64String, 'image/jpeg')
      await sendImage(blob, 'receipt.jpg', TESSERACT_URL)
    } catch (err) {
      if (err?.message === 'User cancelled photos app') {
        setState('idle')
        return
      }
      console.error(err)
      setState('error')
      setErrorMsg(t.scan.error)
    }
  }

  const makeFileHandler = (endpointURL) => async (e) => {
    const file = e.target.files?.[0]
    if (!file) return
    e.target.value = ''
    await sendImage(file, file.name, endpointURL)
  }

  return (
    <div className="scan-page">
      <div className="orb orb-1" />
      <div className="orb orb-2" />

      <div className="scan-card">
        <div className="scan-icon-wrap">
          <div className="scan-ring" />
          <span className="scan-icon">⬡</span>
        </div>

        <h2 className="scan-title">{t.scan.title}</h2>
        <p className="scan-instructions">{t.scan.instructions}</p>

        {state === 'error' && (
          <div className="scan-error">{errorMsg}</div>
        )}

        <div className="scan-actions">
          {state === 'processing' ? (
            <div className="scan-processing">
              <div className="spinner" />
              <span>{t.scan.processing}</span>
            </div>
          ) : (
            <>
              <button className="cta-button" onClick={handleCamera}>
                <span>📷</span>
                {t.scan.openCamera}
              </button>

              <div className="scan-divider">{t.scan.orUpload}</div>

              <button
                className="upload-button upload-button--tesseract"
                onClick={() => tesseractInputRef.current?.click()}
              >
                <span className="upload-button__icon">📄</span>
                <span className="upload-button__label">{t.scan.uploadTesseract}</span>
                <span className="upload-button__badge">OCR</span>
              </button>

              <input
                ref={tesseractInputRef}
                type="file"
                accept="image/*"
                style={{ display: 'none' }}
                onChange={makeFileHandler(TESSERACT_URL)}
              />
            </>
          )}

          <button className="ghost-button" onClick={onCancel} disabled={state === 'processing'}>
            {t.scan.cancel}
          </button>
        </div>
      </div>
    </div>
  )
}

function base64ToBlob(base64, mimeType) {
  const bytes = atob(base64)
  const buffer = new Uint8Array(bytes.length)
  for (let i = 0; i < bytes.length; i++) buffer[i] = bytes.charCodeAt(i)
  return new Blob([buffer], { type: mimeType })
}
