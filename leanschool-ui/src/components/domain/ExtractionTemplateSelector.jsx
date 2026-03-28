import { useState, useEffect } from 'react';
import { useAuth } from '../../auth/useAuth';
import { useTranslation } from '../../i18n/useTranslation';
import './domain-components.css';

export default function ExtractionTemplateSelector({ onSelect, dataVariables, persist = false }) {
  const { t } = useTranslation();
  const { authFetch } = useAuth();
  const [templates, setTemplates] = useState([]);
  const [selectedTemplate, setSelectedTemplate] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  const API = import.meta.env.VITE_EXTRACTION_SERVICE_URL || 'http://localhost:8084';

  // Fetch templates on mount
  useEffect(() => {
    setLoading(true);
    authFetch(`${API}/templates`)
      .then(res => {
        if (!res.ok) throw new Error(`Failed to load templates (${res.status})`);
        return res.json();
      })
      .then(data => setTemplates(data))
      .catch(err => setError(err.message))
      .finally(() => setLoading(false));
  }, [authFetch]);

  const handleSelect = (template) => {
    setSelectedTemplate(template);
    if (persist) {
      triggerExtraction(template);
    } else if (onSelect) {
      onSelect(template.id, dataVariables);
    }
  };

  const triggerExtraction = async (template) => {
    try {
      const extractionRequest = {
        template: template,
        templateData: dataVariables
      };
      
      const res = await authFetch(`${API}/extract`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(extractionRequest)
      });
      
      if (!res.ok) throw new Error(`Extraction failed (${res.status})`);
      
      const blob = await res.blob();
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `extraction-${template.name}.${template.templateType}`;
      a.click();
      URL.revokeObjectURL(url);
    } catch (err) {
      setError(err.message);
    }
  };

  if (loading) return <div className="df-loading">{t.domain.loadingTemplates || 'Loading templates...'}</div>;
  if (error) return <div className="df-error">{error}</div>;

  return (
    <div className="df-template-selector">
      <h3>{t.domain.selectTemplate || 'Select Export Template'}</h3>
      <div className="df-template-list">
        {templates.length === 0 ? (
          <div className="df-no-templates">
            {t.domain.noTemplatesAvailable || 'No templates available'}
          </div>
        ) : (
          templates.map(template => (
            <div 
              key={template.id}
              className={`df-template-card ${selectedTemplate?.id === template.id ? 'selected' : ''}`}
              onClick={() => handleSelect(template)}
            >
              <div className="df-template-name">{template.name}</div>
              <div className="df-template-description">{template.description}</div>
              <div className="df-template-type">{template.templateType}</div>
            </div>
          ))
        )}
      </div>
      {selectedTemplate && persist && (
        <button 
          className="df-export-button" 
          onClick={() => triggerExtraction(selectedTemplate)} 
          disabled={!selectedTemplate}
        >
          {t.domain.exportWithTemplate || 'Export with Selected Template'}
        </button>
      )}
    </div>
  );
}
