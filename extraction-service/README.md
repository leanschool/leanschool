# Extraction Service

A Go service for template-based data extraction supporting Excel, Text, and CSV formats.

## Features

- **Template Management**: Create, update, delete, and retrieve templates
- **Data Extraction**: Fill templates with data using various formats
- **Multiple Format Support**: Excel, Text-based templates, and CSV
- **REST API**: Simple HTTP interface with JSON payloads
- **Authentication**: JWT-based authentication with Keycloak integration

## API Endpoints

### Template Management

- `POST /templates` - Create a new template
- `PUT /templates` - Update an existing template
- `DELETE /templates/{id}` - Delete a template
- `GET /templates` - Get all templates
- `GET /templates/{id}` - Get a specific template

### Data Extraction

- `POST /extract` - Extract data using a template

## Template Types

### Text Template
Uses Go's text/template syntax:
```json
{
  "templateType": "text",
  "template": "Hello {{.Name}}, welcome to {{.School}}!",
  "templateVariables": ["Name", "School"]
}
```

### Excel Template
Uses Excel files with placeholder replacement (simplified implementation):
```json
{
  "templateType": "excel",
  "template": "[binary Excel file content]",
  "templateVariables": ["Variable1", "Variable2"]
}
```

### CSV Template
Generates CSV output with configurable separators and headers:
```json
{
  "templateType": "csv",
  "template": "[CSV configuration]",
  "templateVariables": ["Column1", "Column2"]
}
```

## Authentication

The service uses JWT authentication with the following roles:
- `extraction_read` - Required for GET requests
- `extraction_write` - Required for POST, PUT, DELETE requests

Configure Keycloak integration via environment variables:
- `KEYCLOAK_URL` - Keycloak server URL
- `KEYCLOAK_REALM` - Realm name
- `KEYCLOAK_CLIENT_ID` - Client ID
- `KEYCLOAK_CLIENT_SECRET` - Client secret

## Running the Service

```bash
# Build and run
go build -o extraction-service cmd/main.go
./extraction-service

# Or run directly
go run cmd/main.go
```

The service listens on port `8084` by default.

## Development

### Dependencies

- Go 1.26+
- github.com/xuri/excelize/v2 (for Excel support)

### Building

```bash
go mod tidy
go build ./...
```

### Testing

Run the test script:
```bash
./test_service.sh
```

## Architecture

```
┌─────────────────────────────────────────────────┐
│                 Extraction Service               │
├─────────────────────────────────────────────────┤
│                                                 │
│  ┌─────────────┐    ┌───────────────────┐      │
│  │   Handler   │───▶│ Extraction       │      │
│  └─────────────┘    │   Processor      │      │
│        ▲          └───────────────────┘      │
│        │                  ▲                    │
│  ┌─────┴─────┐      ┌─────┴─────┐              │
│  │  Storage  │      │  Models   │              │
│  └───────────┘      └───────────┘              │
│                                                 │
└─────────────────────────────────────────────────┘
```

## Future Enhancements

- PostgreSQL storage backend
- More sophisticated Excel template processing
- CSV template configuration
- Template versioning
- Caching for frequently used templates
- Rate limiting and request validation
