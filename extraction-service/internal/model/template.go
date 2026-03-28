package model

type TemplateType string

const (
	TemplateTypeExcel    TemplateType = "excel"
	TemplateTypeText     TemplateType = "text"
	TemplateTypeCSV      TemplateType = "csv"
)

type Template struct {
	Id              string      `json:"id"`
	TemplateType    TemplateType   `json:"templateType"`
	Name            string         `json:"name"`
	Description     string         `json:"description"`
	Template        []byte        `json:"template"`
	TemplateVariables []string     `json:"templateVariables"`
}

type ExtractionTemplate struct {
	Id          string          `json:"id"`
	Template    Template            `json:"template"`
	TemplateData map[string]string  `json:"templateData"`
}
