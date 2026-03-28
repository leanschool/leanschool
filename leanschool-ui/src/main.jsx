import React from 'react'
import ReactDOM from 'react-dom/client'
import { defineCustomElements } from '@ionic/pwa-elements/loader';
import App from './App'
import './index.css'

// Define PWA elements (Camera modal, Action sheet, etc.)
defineCustomElements(window);

ReactDOM.createRoot(document.getElementById('root')).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
)