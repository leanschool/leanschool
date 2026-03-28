# Template Management

The template management feature allows school administrators to create, edit, and manage data extraction templates that can be used with the extraction-service. This enables automated data processing from various document formats.

## Accessing Template Management

1. **Prerequisites**: You must have the `school-management` role assigned to your account
2. **Access**: After logging in, you'll see a "Templates" card on your dashboard
3. **Navigation**: Click the "Templates" card to access the template management interface

## Template Types

The system supports three types of templates:

### 1. Text Templates

Text templates use Go's text/template syntax for simple text-based data extraction.

**Use cases**:
- Generating formatted text documents
- Creating email templates
- Simple report generation

**Example**:
```
Name: Student Report Card
Type: Text
Template: Student {{.Name}} received a grade of {{.Grade}} in {{.Subject}}
Variables: Name, Grade, Subject
```

**How it works**:
- Use double curly braces `{{.VariableName}}` to insert variables
- Variables are replaced with actual data during extraction
- Supports basic formatting and text manipulation

### 2. Excel Templates

Excel templates allow you to create structured spreadsheets with placeholder values.

**Use cases**:
- Budget reports
- Student grade sheets
- Financial statements
- Data export templates

**Example**:
```
Name: Monthly Budget Report
Type: Excel
Template: [Upload Excel file with placeholders]
Variables: Month, Income, Expenses, Balance
```

**How to create**:
1. Create an Excel file with your desired structure
2. Use specific cell formatting or naming conventions for placeholders
3. Upload the Excel file as the template
4. Define the variables that will be replaced

**Note**: The current implementation supports basic placeholder replacement. Complex Excel formulas and formatting are preserved.

### 3. CSV Templates

CSV templates enable structured data export in comma-separated values format.

**Use cases**:
- Data imports/exports
- Simple tabular data representation
- System integration

**Example**:
```
Name: Student Contact List
Type: CSV
Template: Name,Email,Phone,Class
Variables: Name, Email, Phone, Class
```

**How it works**:
- Define column headers in the template
- Variables represent the data that will populate each row
- During extraction, multiple data records can be processed

## Creating a New Template

1. **Click "New Template"**: Located at the bottom of the templates list
2. **Fill in template details**:
   - **Name**: A descriptive name for your template
   - **Type**: Select from Text, Excel, or CSV
   - **Template**: The actual template content
   - **Variables**: Comma-separated list of variables used in the template

3. **Save**: Click "Create" to save your template

## Editing a Template

1. **Find your template**: Locate it in the templates list
2. **Click the edit button** (✎ icon)
3. **Make your changes**: Update any field as needed
4. **Save changes**: Click "Save" to apply your updates

## Deleting a Template

1. **Find the template**: Locate it in the templates list
2. **Click the delete button** (✕ icon)
3. **Confirm**: The template will be permanently removed

## Template Variables

Variables are the dynamic parts of your templates that get replaced with actual data during extraction.

**Best practices**:
- Use descriptive variable names (e.g., `StudentName` instead of `name1`)
- Keep variable names consistent across templates
- Use camelCase or snake_case for multi-word variables
- Document what each variable represents

**Variable format**:
- Comma-separated list in the variables field
- Example: `Name, Date, Amount, Description`
- Spaces around commas are automatically trimmed

## Using Templates with the Extraction Service

Once created, templates can be used programmatically via the extraction-service API:

```bash
# Get all templates
GET /templates

# Get a specific template
GET /templates/{id}

# Use a template for extraction
POST /extract
{
  "templateId": "your-template-id",
  "data": {
    "Name": "John Doe",
    "Grade": "A",
    "Subject": "Mathematics"
  }
}
```

## Advanced Tips

### Template Organization
- Use consistent naming conventions (e.g., `Budget_2024_Q1`, `Student_Report_Card`)
- Group related templates by type or purpose
- Include the template type in the name for easy identification

### Testing Templates
- Start with simple templates and gradually add complexity
- Test with sample data before using in production
- Validate that all variables are properly replaced

### Performance Considerations
- Complex Excel templates with many formulas may process slower
- Large CSV templates with many columns should be optimized
- Text templates are generally the fastest to process

## Troubleshooting

**Common issues and solutions**:

1. **Template not appearing**: 
   - Check that you have the `school-management` role
   - Verify the template was saved successfully
   - Refresh the page if needed

2. **Variable replacement not working**:
   - Ensure variable names match exactly (case-sensitive)
   - Check for typos in variable names
   - Verify the template type supports the syntax you're using

3. **Excel template issues**:
   - Make sure the Excel file is not corrupted
   - Check that placeholder cells are properly formatted
   - Complex formulas may need simplification

4. **API errors**:
   - Verify the extraction-service is running
   - Check your API authentication credentials
   - Review error messages for specific issues

## Security Considerations

- Templates are only accessible to users with the `school-management` role
- Sensitive data should not be stored in template content
- Regularly review and clean up unused templates
- Use descriptive names that don't reveal sensitive information

## Integration with Other Systems

Templates can be used in various workflows:

1. **Automated reporting**: Schedule template-based reports
2. **Data import/export**: Use CSV templates for system integration
3. **Document generation**: Create standardized documents from templates
4. **Email automation**: Generate personalized emails using text templates

## Future Enhancements

The template system is designed to be extensible. Future improvements may include:
- Template versioning and history
- Template sharing between organizations
- Advanced Excel formula support
- PDF template support
- Template validation tools
- Usage analytics and reporting