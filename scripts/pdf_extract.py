#!/usr/bin/env python3
"""
PDF extraction tool optimized for Korean documents
Using pdfplumber for high-quality text and table extraction
"""

import json
import sys
import argparse
from pathlib import Path
from typing import Dict, List, Any

try:
    import pdfplumber
except ImportError:
    print(json.dumps({
        "error": "pdfplumber not installed",
        "message": "Please install pdfplumber: pip install pdfplumber"
    }))
    sys.exit(1)


class PDFExtractor:
    """PDF extractor optimized for Korean documents"""
    
    def __init__(self, filepath: str, debug: bool = False):
        self.filepath = filepath
        self.debug = debug
        
    def extract(self) -> Dict[str, Any]:
        """Extract content from PDF"""
        result = {
            "success": True,
            "filename": Path(self.filepath).name,
            "pages": [],
            "metadata": {},
            "error": None
        }
        
        try:
            with pdfplumber.open(self.filepath) as pdf:
                # Extract metadata
                if pdf.metadata:
                    result["metadata"] = {
                        "title": pdf.metadata.get("Title", ""),
                        "author": pdf.metadata.get("Author", ""),
                        "subject": pdf.metadata.get("Subject", ""),
                        "creator": pdf.metadata.get("Creator", ""),
                        "total_pages": len(pdf.pages)
                    }
                
                # Process each page
                for page_num, page in enumerate(pdf.pages, 1):
                    if self.debug:
                        print(f"Processing page {page_num}...", file=sys.stderr)
                    
                    page_data = self.extract_page(page, page_num)
                    result["pages"].append(page_data)
                    
        except Exception as e:
            result["success"] = False
            result["error"] = str(e)
            if self.debug:
                import traceback
                traceback.print_exc(file=sys.stderr)
        
        return result
    
    def extract_page(self, page, page_num: int) -> Dict[str, Any]:
        """Extract content from a single page"""
        page_data = {
            "number": page_num,
            "text": "",
            "tables": [],
            "layout": {
                "width": float(page.width) if page.width else 0,
                "height": float(page.height) if page.height else 0
            }
        }
        
        # Extract text
        text = page.extract_text()
        if text:
            page_data["text"] = text
        
        # Extract tables with optimized settings for Korean documents
        tables = self.extract_tables(page)
        page_data["tables"] = tables
        
        return page_data
    
    def extract_tables(self, page) -> List[Dict[str, Any]]:
        """Extract tables with multiple strategies for Korean documents"""
        tables = []
        
        # Try multiple strategies to find tables
        strategies = [
            # Strategy 1: Lines-based (for tables with visible borders)
            {
                "vertical_strategy": "lines",
                "horizontal_strategy": "lines",
                "snap_tolerance": 3,
                "join_tolerance": 3,
                "edge_min_length": 3,
                "min_words_vertical": 0,
                "min_words_horizontal": 0,
            },
            # Strategy 2: Text-based (for tables without clear borders)
            {
                "vertical_strategy": "text",
                "horizontal_strategy": "text",
                "snap_tolerance": 5,
                "join_tolerance": 5,
                "edge_min_length": 5,
                "min_words_vertical": 0,
                "min_words_horizontal": 0,
                "text_tolerance": 3,
                "text_x_tolerance": 5,
                "text_y_tolerance": 3,
            },
            # Strategy 3: Mixed (lines for vertical, text for horizontal)
            {
                "vertical_strategy": "lines",
                "horizontal_strategy": "text",
                "snap_tolerance": 3,
                "join_tolerance": 3,
                "edge_min_length": 3,
                "min_words_vertical": 0,
                "min_words_horizontal": 0,
            }
        ]
        
        best_tables = []
        best_score = 0
        
        for strategy_idx, table_settings in enumerate(strategies):
            try:
                # Find tables with current strategy
                page_tables = page.find_tables(table_settings)
                
                if self.debug and page_tables:
                    print(f"  Strategy {strategy_idx + 1} found {len(page_tables)} table(s)", file=sys.stderr)
                
                current_tables = []
                current_score = 0
                
                for table in page_tables:
                    # Extract table data
                    extracted = table.extract()
                    
                    if extracted and len(extracted) > 0:
                        # Calculate quality score
                        non_empty_cells = sum(1 for row in extracted for cell in row if cell and str(cell).strip())
                        total_cells = len(extracted) * (len(extracted[0]) if extracted else 0)
                        
                        if total_cells > 0:
                            fill_ratio = non_empty_cells / total_cells
                            # Score based on fill ratio and table size
                            score = fill_ratio * total_cells
                            current_score += score
                            
                            current_tables.append({
                                "table": table,
                                "data": extracted,
                                "score": score
                            })
                
                # Keep the best strategy results
                if current_score > best_score:
                    best_score = current_score
                    best_tables = current_tables
                    if self.debug:
                        print(f"    Strategy {strategy_idx + 1} is current best with score {current_score:.2f}", file=sys.stderr)
                        
            except Exception as e:
                if self.debug:
                    print(f"  Strategy {strategy_idx + 1} error: {e}", file=sys.stderr)
        
        # Process the best tables found
        for table_info in best_tables:
            table = table_info["table"]
            extracted = table_info["data"]
            
            # Clean and process table data
            cleaned_table = self.clean_table(extracted)
            
            if cleaned_table:
                # Get table position
                bbox = table.bbox
                
                table_data = {
                    "index": len(tables),
                    "data": cleaned_table,
                    "rows": len(cleaned_table),
                    "cols": len(cleaned_table[0]) if cleaned_table else 0,
                    "position": {
                        "x": float(bbox[0]),
                        "y": float(bbox[1]),
                        "width": float(bbox[2] - bbox[0]),
                        "height": float(bbox[3] - bbox[1])
                    }
                }
                tables.append(table_data)
                
                if self.debug:
                    print(f"  Selected table: {table_data['rows']}x{table_data['cols']}", file=sys.stderr)
        
        return tables
    
    def clean_table(self, table_data: List[List]) -> List[List[str]]:
        """Clean and normalize table data"""
        if not table_data:
            return []
        
        cleaned = []
        for row in table_data:
            cleaned_row = []
            for cell in row:
                if cell is None:
                    cleaned_row.append("")
                else:
                    # Handle multi-line cells
                    cell_text = str(cell).strip()
                    # Preserve newlines for layout
                    cleaned_row.append(cell_text)
            cleaned.append(cleaned_row)
        
        # Remove completely empty rows
        cleaned = [row for row in cleaned if any(cell for cell in row)]
        
        return cleaned


def main():
    parser = argparse.ArgumentParser(
        description="Extract text and tables from PDF files"
    )
    parser.add_argument("pdf_file", help="Path to PDF file")
    parser.add_argument("--debug", "-d", action="store_true", 
                       help="Show debug information")
    parser.add_argument("--pretty", "-p", action="store_true",
                       help="Pretty print JSON output")
    parser.add_argument("--output", "-o", help="Output file (default: stdout)")
    
    args = parser.parse_args()
    
    # Check if file exists
    if not Path(args.pdf_file).exists():
        result = {
            "success": False,
            "error": f"File not found: {args.pdf_file}"
        }
        print(json.dumps(result))
        sys.exit(1)
    
    # Extract PDF content
    extractor = PDFExtractor(args.pdf_file, debug=args.debug)
    result = extractor.extract()
    
    # Output result
    if args.pretty:
        json_output = json.dumps(result, ensure_ascii=False, indent=2)
    else:
        json_output = json.dumps(result, ensure_ascii=False)
    
    if args.output:
        with open(args.output, 'w', encoding='utf-8') as f:
            f.write(json_output)
        if args.debug:
            print(f"Output written to: {args.output}", file=sys.stderr)
    else:
        print(json_output)


if __name__ == "__main__":
    main()