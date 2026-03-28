package handler

import (
	"bytes"
	"encoding/csv"
	"errors"
	"text/template"

	"github.com/Joel-Haeberli/extraction-service/internal/model"
	"github.com/xuri/excelize/v2"
)

type ExtractionProcessor struct{}

func NewExtractionProcessor() *ExtractionProcessor {
	return &ExtractionProcessor{}
}

func (p *ExtractionProcessor) Process(extraction model.ExtractionTemplate) ([]byte, error) {
	switch extraction.Template.TemplateType {
	case model.TemplateTypeExcel:
		return p.processExcel(extraction)
	case model.TemplateTypeText:
		return p.processText(extraction)
	case model.TemplateTypeCSV:
		return p.processCSV(extraction)
	default:
		return nil, errors.New("unsupported template type")
	}
}

func (p *ExtractionProcessor) processExcel(extraction model.ExtractionTemplate) ([]byte, error) {
	f, err := excelize.OpenReader(bytes.NewReader(extraction.Template.Template))
	if err != nil {
		return nil, err
	}

	// Apply template data to Excel file
	for _, sheetName := range f.GetSheetList() {
		rows, err := f.GetRows(sheetName)
		if err != nil {
			return nil, err
		}

		for _, row := range rows {
			for _, cell := range row {
				// Replace placeholders in the format {{.VariableName}}
				for varName := range extraction.TemplateData {
					if len(cell) >= len(varName)+4 && cell == "{{"+varName+"}}" {
						// This is a simplified approach - in a real implementation,
						// you would need to parse the Excel file more carefully
						// and replace placeholders in all cells
					}
				}
			}
		}
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (p *ExtractionProcessor) processText(extraction model.ExtractionTemplate) ([]byte, error) {
	tmpl, err := template.New("extraction").Parse(string(extraction.Template.Template))
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, extraction.TemplateData); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (p *ExtractionProcessor) processCSV(extraction model.ExtractionTemplate) ([]byte, error) {
	// Parse CSV template to get separator and header
	// This is a simplified implementation - in a real scenario,
	// the template would contain CSV-specific metadata
	
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	defer writer.Flush()

	// Write header (this would come from template metadata in a real implementation)
	if err := writer.Write([]string{"Column1", "Column2"}); err != nil {
		return nil, err
	}

	// Write data row with template data
	row := make([]string, 2)
	for i, varName := range extraction.Template.TemplateVariables {
		if i < 2 { // Only write first 2 variables for this simplified example
			if value, exists := extraction.TemplateData[varName]; exists {
				row[i] = value
			}
		}
	}

	if err := writer.Write(row); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
