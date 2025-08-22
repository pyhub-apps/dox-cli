#!/bin/bash

# Batch Document Processing Script
# This script demonstrates how to use dox for complex document workflows

set -e  # Exit on error

echo "==================================="
echo "Document Batch Processing Script"
echo "==================================="

# Configuration
DOCS_DIR="./documents"
BACKUP_DIR="./backups"
OUTPUT_DIR="./output"
REPORTS_DIR="./reports"

# Create directories if they don't exist
mkdir -p "$BACKUP_DIR" "$OUTPUT_DIR" "$REPORTS_DIR"

# Step 1: Create backup of all documents
echo ""
echo "Step 1: Creating backups..."
echo "---------------------------"
cp -r "$DOCS_DIR" "$BACKUP_DIR/$(date +%Y%m%d_%H%M%S)"
echo "✅ Backup created"

# Step 2: Update year in all documents
echo ""
echo "Step 2: Updating year references..."
echo "------------------------------------"
cat > /tmp/year-update.yml << EOF
- old: "2024"
  new: "2025"
- old: "FY2024"
  new: "FY2025"
- old: "2024-"
  new: "2025-"
EOF

dox replace --rules /tmp/year-update.yml --path "$DOCS_DIR" --concurrent
echo "✅ Year updated"

# Step 3: Update company information
echo ""
echo "Step 3: Updating company information..."
echo "----------------------------------------"
cat > /tmp/company-update.yml << EOF
- old: "OldCompany Inc."
  new: "NewCompany Ltd."
- old: "support@oldcompany.com"
  new: "support@newcompany.com"
EOF

dox replace --rules /tmp/company-update.yml --path "$DOCS_DIR" --backup
echo "✅ Company information updated"

# Step 4: Generate changelog from markdown
echo ""
echo "Step 4: Generating changelog document..."
echo "-----------------------------------------"
cat > /tmp/changelog.md << EOF
# Changes Applied - $(date +"%B %d, %Y")

## Document Updates
- Updated year references from 2024 to 2025
- Changed company name to NewCompany Ltd.
- Updated contact email addresses

## Files Modified
- $(find "$DOCS_DIR" -name "*.docx" -o -name "*.pptx" | wc -l) documents processed

## Next Steps
- Review updated documents
- Distribute to stakeholders
- Archive old versions
EOF

dox create --from /tmp/changelog.md --output "$REPORTS_DIR/changelog.docx"
echo "✅ Changelog created"

# Step 5: Process templates with current date
echo ""
echo "Step 5: Processing templates..."
echo "--------------------------------"
if [ -f "templates/report-template.docx" ]; then
    dox template \
        --template templates/report-template.docx \
        --output "$OUTPUT_DIR/report-$(date +%Y%m%d).docx" \
        --set date="$(date +"%B %d, %Y")" \
        --set author="$USER" \
        --set version="2.0"
    echo "✅ Report template processed"
fi

# Step 6: Generate AI summary (if API key is set)
echo ""
echo "Step 6: Generating AI summary..."
echo "---------------------------------"
if [ -n "$OPENAI_API_KEY" ]; then
    dox generate --type report \
        --prompt "Summarize the document processing activities completed today" \
        --output "$REPORTS_DIR/summary.md"
    
    # Convert summary to Word
    dox create --from "$REPORTS_DIR/summary.md" --output "$REPORTS_DIR/summary.docx"
    echo "✅ AI summary generated"
else
    echo "⚠️  Skipped: OPENAI_API_KEY not set"
fi

# Step 7: Create final report
echo ""
echo "Step 7: Creating final report..."
echo "---------------------------------"
cat > /tmp/final-report.md << EOF
# Batch Processing Complete

**Date**: $(date +"%B %d, %Y %H:%M")  
**Processed by**: $USER  
**Documents processed**: $(find "$DOCS_DIR" -name "*.docx" -o -name "*.pptx" | wc -l)  

## Summary
All documents have been successfully processed with the following updates:
- Year references updated to 2025
- Company information updated
- Templates processed with current data
- Backups created for all documents

## Output Locations
- Updated documents: $DOCS_DIR
- Backups: $BACKUP_DIR
- Reports: $REPORTS_DIR
- Generated files: $OUTPUT_DIR

---
*This report was automatically generated using dox document automation tool*
EOF

dox create --from /tmp/final-report.md --output "$REPORTS_DIR/processing-report.docx"
echo "✅ Final report created"

# Cleanup temporary files
rm -f /tmp/year-update.yml /tmp/company-update.yml /tmp/changelog.md /tmp/final-report.md

echo ""
echo "==================================="
echo "✨ Batch processing complete!"
echo "==================================="
echo ""
echo "Results:"
echo "  📁 Updated documents: $DOCS_DIR"
echo "  💾 Backups: $BACKUP_DIR"
echo "  📊 Reports: $REPORTS_DIR"
echo "  📄 Output files: $OUTPUT_DIR"
echo ""