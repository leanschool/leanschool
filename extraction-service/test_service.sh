#!/bin/bash

# Build the service
echo "Building extraction service..."
go build -o extraction-service cmd/main.go

# Start the service in the background
echo "Starting extraction service..."
./extraction-service &
SERVICE_PID=$!

# Wait a moment for the service to start
sleep 2

# Test creating a template
echo "Testing template creation..."
curl -X POST http://localhost:8084/templates \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Template",
    "description": "A test template",
    "templateType": "text",
    "template": "Hello {{.Name}}, welcome to {{.School}}!",
    "templateVariables": ["Name", "School"]
  }'

echo ""
echo "Testing template extraction..."

# Test extraction
curl -X POST http://localhost:8084/extract \
  -H "Content-Type: application/json" \
  -d '{
    "template": {
      "name": "Test Template",
      "description": "A test template",
      "templateType": "text",
      "template": "Hello {{.Name}}, welcome to {{.School}}!",
      "templateVariables": ["Name", "School"]
    },
    "templateData": {
      "Name": "John Doe",
      "School": "Leanschool"
    }
  }'

echo ""
echo "Testing get all templates..."
curl -X GET http://localhost:8084/templates

# Clean up
echo ""
echo "Stopping service..."
kill $SERVICE_PID
rm -f extraction-service
