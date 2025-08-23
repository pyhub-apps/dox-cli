# Templates Guide

Learn how to create and use document templates with dox for automated document generation.

## 📋 Overview

Templates allow you to create reusable document layouts with variable placeholders that can be filled with data from YAML or JSON files.

## 🎯 Template Basics

### What are Templates?

Templates are Word (.docx) or PowerPoint (.pptx) documents containing:
- Static content (headers, footers, boilerplate text)
- Variable placeholders (`{{variable_name}}`)
- Conditional sections
- Repeating elements (lists, tables)

### Variable Syntax

Basic variable replacement:
```
Dear {{customer_name}},

Your order #{{order_number}} has been confirmed.
Total: {{total_amount}}

Thank you for your business!
```

## 📝 Creating Templates

### Word Templates

1. **Create a new Word document**
2. **Design your layout** with styles, fonts, and formatting
3. **Insert placeholders** where dynamic content should appear
4. **Save as .docx**

Example Word template content:
```
INVOICE
================
Invoice #: {{invoice_number}}
Date: {{invoice_date}}

Bill To:
{{client_name}}
{{client_address}}

Items:
{{range items}}
- {{.description}}: {{.amount}}
{{end}}

Subtotal: {{subtotal}}
Tax: {{tax}}
Total: {{total}}
```

### PowerPoint Templates

1. **Create a presentation** with your design
2. **Add placeholders** in text boxes
3. **Use master slides** for consistent formatting
4. **Save as .pptx**

Example PowerPoint slide:
```
[Slide 1 - Title]
{{presentation_title}}
{{presenter_name}}
{{date}}

[Slide 2 - Agenda]
Today's Topics:
{{range topics}}
• {{.title}}
{{end}}

[Slide 3 - Details]
{{content}}
```

## 🔧 Advanced Template Features

### Conditional Content

Show content based on conditions:
```
{{if premium_customer}}
Thank you for being a Premium member!
You enjoy exclusive benefits:
- Free shipping
- 24/7 support
- 20% discount
{{end}}

{{if not paid}}
PAYMENT DUE: Please pay by {{due_date}}
{{end}}
```

### Loops and Lists

Iterate over arrays:
```
Order Items:
{{range items}}
Product: {{.name}}
Quantity: {{.quantity}}
Price: {{.price}}
Subtotal: {{.subtotal}}
---
{{end}}
```

### Nested Variables

Access nested data structures:
```
Company: {{company.name}}
Address: {{company.address.street}}
City: {{company.address.city}}
Contact: {{company.contact.email}}
```

### Formatting Functions

Apply formatting to variables:
```
Date: {{formatDate invoice_date "January 2, 2006"}}
Amount: {{formatCurrency amount "USD"}}
Percentage: {{formatPercent rate}}
Upper: {{upper company_name}}
Lower: {{lower email}}
```

## 📊 Data Files

### YAML Format

```yaml
# invoice-data.yml
invoice_number: "INV-2025-001"
invoice_date: "2025-01-15"
client_name: "Acme Corporation"
client_address: "123 Business St, New York, NY 10001"

items:
  - description: "Consulting Services"
    amount: "$5,000"
  - description: "Software License"
    amount: "$2,000"
  - description: "Support Package"
    amount: "$1,000"

subtotal: "$8,000"
tax: "$800"
total: "$8,800"

premium_customer: true
paid: false
due_date: "2025-02-15"
```

### JSON Format

```json
{
  "invoice_number": "INV-2025-001",
  "invoice_date": "2025-01-15",
  "client_name": "Acme Corporation",
  "client_address": "123 Business St, New York, NY 10001",
  "items": [
    {
      "description": "Consulting Services",
      "amount": "$5,000"
    },
    {
      "description": "Software License",
      "amount": "$2,000"
    }
  ],
  "subtotal": "$8,000",
  "tax": "$800",
  "total": "$8,800",
  "premium_customer": true,
  "paid": false,
  "due_date": "2025-02-15"
}
```

## 🚀 Using Templates

### Basic Usage

```bash
# Fill template with YAML data
dox template --template invoice.docx --values invoice-data.yml --output invoice-final.docx

# Fill template with JSON data
dox template -t report.pptx -v data.json -o report-final.pptx
```

### Override Values

```bash
# Override specific values from command line
dox template -t letter.docx -v base.yml -o output.docx \
  --set "date=2025-01-20" \
  --set "signature=John Doe"
```

### Multiple Data Files

```bash
# Combine multiple data sources (later files override earlier ones)
dox template -t contract.docx \
  -v common.yml \
  -v client-specific.yml \
  -v overrides.yml \
  -o contract-final.docx
```

## 📋 Template Examples

### Invoice Template

**Template (invoice-template.docx):**
```
                    INVOICE
        
Company: {{company.name}}
Address: {{company.address}}
Phone: {{company.phone}}

Bill To:
{{client.name}}
{{client.address}}

Invoice #: {{invoice.number}}
Date: {{invoice.date}}
Due Date: {{invoice.due_date}}

┌─────────────────────────────────────────┐
│ Description          │ Qty │ Rate │ Amount │
├─────────────────────────────────────────┤
{{range items}}
│ {{.description}} │ {{.quantity}} │ {{.rate}} │ {{.amount}} │
{{end}}
├─────────────────────────────────────────┤
│ Subtotal:                    │ {{totals.subtotal}} │
│ Tax ({{totals.tax_rate}}%): │ {{totals.tax}} │
│ TOTAL:                       │ {{totals.total}} │
└─────────────────────────────────────────┘

{{if notes}}
Notes: {{notes}}
{{end}}
```

### Contract Template

**Template (contract-template.docx):**
```
SERVICE AGREEMENT

This Agreement is entered into as of {{contract.date}} between:

{{party_a.name}} ("Service Provider")
{{party_a.address}}

AND

{{party_b.name}} ("Client")  
{{party_b.address}}

SERVICES:
{{range services}}
{{.number}}. {{.description}}
{{end}}

PAYMENT:
Amount: {{payment.amount}}
Terms: {{payment.terms}}

DURATION:
Start Date: {{duration.start}}
End Date: {{duration.end}}

{{if special_terms}}
SPECIAL TERMS:
{{special_terms}}
{{end}}

SIGNATURES:
_________________     _________________
{{party_a.representative}}    {{party_b.representative}}
{{party_a.title}}           {{party_b.title}}
Date: _________       Date: _________
```

### Report Template

**Template (report-template.pptx):**
```
[Slide 1]
{{report.title}}
{{report.subtitle}}
{{report.date}}

[Slide 2]
Executive Summary
{{summary}}

[Slide 3]
Key Metrics
{{range metrics}}
• {{.name}}: {{.value}} ({{.change}})
{{end}}

[Slide 4]
Recommendations
{{range recommendations}}
{{.priority}}. {{.action}}
   Impact: {{.impact}}
   Timeline: {{.timeline}}
{{end}}

[Slide 5]
Next Steps
{{next_steps}}

Contact: {{author.name}}
Email: {{author.email}}
```

## 🎨 Styling Templates

### Preserving Formatting

- Template formatting is preserved in output
- Styles, fonts, colors remain intact
- Only placeholder text is replaced

### Dynamic Styling

Use conditional formatting:
```
{{if status == "overdue"}}
<text color="red">OVERDUE</text>
{{else}}
<text color="green">Current</text>
{{end}}
```

### Tables and Lists

Templates can include:
- Formatted tables with variables
- Numbered/bulleted lists
- Charts (with static data)
- Images and logos

## 🔍 Troubleshooting

### Common Issues

#### Variables Not Replaced
- Check variable names match exactly (case-sensitive)
- Ensure data file has correct structure
- Verify YAML/JSON syntax is valid

#### Formatting Lost
- Use proper template document (not plain text)
- Ensure placeholders are in text content, not fields
- Save template in .docx/.pptx format

#### Loops Not Working
- Check array structure in data file
- Use correct range syntax
- Ensure proper end tags

### Validation

Test your template:
```bash
# Dry run to check for errors
dox template -t template.docx -v data.yml -o test.docx --dry-run

# Validate data file
dox template --validate-data data.yml

# Check for missing variables
dox template -t template.docx -v data.yml -o out.docx --missing=warn
```

## 📚 Best Practices

### Template Design
1. **Keep it simple** - Start with basic variables
2. **Test incrementally** - Add features gradually
3. **Document variables** - List all required variables
4. **Use meaningful names** - `invoice_number` not `var1`
5. **Provide defaults** - Handle missing data gracefully

### Data Organization
1. **Structure logically** - Group related data
2. **Use consistent types** - Don't mix strings and numbers
3. **Validate input** - Check data before processing
4. **Version control** - Track template and data changes

### Performance
1. **Optimize loops** - Minimize nested iterations
2. **Cache templates** - Reuse for multiple documents
3. **Batch processing** - Process multiple documents together

## 🎯 Real-World Examples

### Monthly Reports
```bash
# Generate monthly reports for all clients
for client in clients/*.yml; do
  dox template -t monthly-report.docx -v "$client" \
    -o "reports/$(basename $client .yml)-report.docx"
done
```

### Batch Invoices
```bash
# Create invoices from CSV data
dox template --template invoice.docx \
  --values invoices.csv \
  --output-pattern "invoices/invoice-{invoice_number}.docx" \
  --batch
```

### Personalized Letters
```bash
# Mail merge functionality
dox template -t letter.docx -v recipients.yml \
  --output-pattern "letters/{name}-letter.docx" \
  --personalize
```

## 📖 Next Steps

- Explore [AI Generation](./ai-generation.md) for dynamic content
- See [Command Reference](./commands.md) for all options
- Check [Examples](../examples/) for more use cases