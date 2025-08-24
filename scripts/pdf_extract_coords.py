#!/usr/bin/env python3
"""
PDF extraction with coordinate-based structure preservation
Focuses on text layer only, preserving spatial layout
"""

import json
import sys
import argparse
from pathlib import Path
from typing import Dict, List, Any, Tuple
import pdfplumber


class CoordinateBasedExtractor:
    """Extract PDF text with coordinate-based structure detection"""
    
    def __init__(self, filepath: str, debug: bool = False, 
                 min_quality: float = 0.2, strict: bool = False):
        self.filepath = filepath
        self.debug = debug
        self.min_quality = min_quality  # Minimum acceptable quality (0.0-1.0)
        self.strict = strict  # Strict quality mode
        self.quality_report = {
            "overall_score": 0.0,
            "table_scores": [],
            "warnings": [],
            "errors": []
        }
        
    def extract(self) -> Dict[str, Any]:
        """Extract content preserving spatial structure"""
        result = {
            "success": True,
            "filename": Path(self.filepath).name,
            "pages": [],
            "metadata": {},
            "error": None,
            "quality": None  # Will be added after quality assessment
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
                    
                    page_data = self.extract_page_with_coords(page, page_num)
                    result["pages"].append(page_data)
                
                # Assess quality
                quality_score = self.assess_quality(result)
                result["quality"] = {
                    "score": quality_score,
                    "report": self.quality_report
                }
                
                # Check quality threshold
                if quality_score < self.min_quality:
                    error_msg = f"Quality too low: {quality_score:.2f} (minimum: {self.min_quality:.2f})"
                    if self.strict:
                        result["success"] = False
                        result["error"] = error_msg
                        self.quality_report["errors"].append(error_msg)
                    else:
                        self.quality_report["warnings"].append(error_msg)
                        
        except Exception as e:
            result["success"] = False
            result["error"] = str(e)
            if self.debug:
                import traceback
                traceback.print_exc(file=sys.stderr)
        
        return result
    
    def extract_page_with_coords(self, page, page_num: int) -> Dict[str, Any]:
        """Extract page content with coordinate-based structure"""
        page_data = {
            "number": page_num,
            "layout": {
                "width": float(page.width) if page.width else 0,
                "height": float(page.height) if page.height else 0
            },
            "elements": []  # List of structured elements with coordinates
        }
        
        # Get all text with coordinates
        chars = page.chars if page.chars else []
        
        if self.debug:
            print(f"  Page {page_num}: {len(chars)} characters", file=sys.stderr)
        
        # Group text by lines based on Y coordinate
        lines = self.group_text_by_lines(chars)
        
        # Detect structure (headings, lists, tables)
        structured_elements = self.detect_structure(lines, page)
        
        page_data["elements"] = structured_elements
        
        # Also extract tables using pdfplumber
        tables = self.extract_tables_with_coords(page)
        if tables:
            page_data["tables"] = tables
        
        return page_data
    
    def group_text_by_lines(self, chars: List[Dict]) -> List[Dict]:
        """Group characters into lines based on Y coordinate"""
        if not chars:
            return []
        
        # Group by Y coordinate (with tolerance)
        y_tolerance = 2
        lines = []
        current_line = None
        
        # Sort by Y then X
        sorted_chars = sorted(chars, key=lambda c: (round(c['top']), c['x0']))
        
        for char in sorted_chars:
            y = char['top']
            
            if current_line is None or abs(y - current_line['y']) > y_tolerance:
                # Start new line
                if current_line:
                    lines.append(current_line)
                current_line = {
                    'y': y,
                    'x_min': char['x0'],
                    'x_max': char['x1'],
                    'height': char['height'],
                    'chars': [char],
                    'text': char['text']
                }
            else:
                # Add to current line
                current_line['chars'].append(char)
                current_line['text'] += char['text']
                current_line['x_min'] = min(current_line['x_min'], char['x0'])
                current_line['x_max'] = max(current_line['x_max'], char['x1'])
        
        if current_line:
            lines.append(current_line)
        
        return lines
    
    def detect_structure(self, lines: List[Dict], page) -> List[Dict]:
        """Detect document structure from lines"""
        elements = []
        
        for line in lines:
            text = line['text'].strip()
            if not text:
                continue
            
            element = {
                "type": "text",  # default
                "content": text,
                "bbox": {
                    "x": float(line['x_min']),
                    "y": float(line['y']),
                    "width": float(line['x_max'] - line['x_min']),
                    "height": float(line['height'])
                }
            }
            
            # Detect heading (short lines, specific patterns)
            if self.is_heading(text, line, page):
                element["type"] = "heading"
                element["level"] = self.get_heading_level(text, line)
            
            # Detect list items
            elif self.is_list_item(text):
                element["type"] = "list_item"
                element["marker"] = self.get_list_marker(text)
            
            # Detect table-like structure (based on alignment)
            elif self.is_table_row(line, lines):
                element["type"] = "table_row"
            
            elements.append(element)
        
        return elements
    
    def is_heading(self, text: str, line: Dict, page) -> bool:
        """Detect if text is a heading"""
        # Korean document heading patterns
        if text.startswith('▢') or text.startswith('■'):
            return True
        
        # Page markers
        if text.startswith('- ') and text.endswith(' -') and len(text) < 10:
            return True
        
        # Short centered text
        if len(text) < 50:
            page_center = page.width / 2 if page.width else 300
            line_center = (line['x_min'] + line['x_max']) / 2
            if abs(line_center - page_center) < 50:  # Roughly centered
                return True
        
        return False
    
    def get_heading_level(self, text: str, line: Dict) -> int:
        """Determine heading level"""
        if text.startswith('▢') or text.startswith('■'):
            return 1
        elif text.startswith('- ') and text.endswith(' -'):
            return 3
        else:
            return 2
    
    def is_list_item(self, text: str) -> bool:
        """Detect if text is a list item"""
        markers = ['○', '●', '▪', '▫', '◦', '•', '-', '*']
        return any(text.startswith(marker + ' ') or text == marker for marker in markers)
    
    def get_list_marker(self, text: str) -> str:
        """Extract list marker"""
        markers = ['○', '●', '▪', '▫', '◦', '•', '-', '*']
        for marker in markers:
            if text.startswith(marker):
                return marker
        return ""
    
    def is_table_row(self, line: Dict, all_lines: List[Dict]) -> bool:
        """Detect if line might be part of a table"""
        # Check if multiple items are aligned vertically
        y = line['y']
        similar_y_lines = [l for l in all_lines if abs(l['y'] - y) < 2]
        
        # If multiple text segments at same Y, might be table
        if len(similar_y_lines) > 1:
            return True
        
        return False
    
    def extract_tables_with_coords(self, page) -> List[Dict[str, Any]]:
        """Extract tables with coordinate information"""
        tables = []
        
        # Multiple strategies for table detection
        strategies = [
            {
                "vertical_strategy": "lines",
                "horizontal_strategy": "lines",
                "snap_tolerance": 3,
                "join_tolerance": 3,
                "edge_min_length": 3,
            },
            {
                "vertical_strategy": "text", 
                "horizontal_strategy": "text",
                "snap_tolerance": 5,
                "join_tolerance": 5,
            }
        ]
        
        for strategy in strategies:
            try:
                page_tables = page.find_tables(strategy)
                
                for table in page_tables:
                    extracted = table.extract()
                    
                    if extracted and len(extracted) > 0:
                        # Check if table has content
                        non_empty = sum(1 for row in extracted for cell in row if cell and str(cell).strip())
                        
                        if non_empty > 0:
                            bbox = table.bbox
                            table_data = {
                                "data": extracted,
                                "rows": len(extracted),
                                "cols": len(extracted[0]) if extracted else 0,
                                "bbox": {
                                    "x": float(bbox[0]),
                                    "y": float(bbox[1]),
                                    "width": float(bbox[2] - bbox[0]),
                                    "height": float(bbox[3] - bbox[1])
                                }
                            }
                            
                            # Don't add duplicate tables
                            is_duplicate = False
                            for existing in tables:
                                if (abs(existing["bbox"]["x"] - table_data["bbox"]["x"]) < 10 and
                                    abs(existing["bbox"]["y"] - table_data["bbox"]["y"]) < 10):
                                    is_duplicate = True
                                    break
                            
                            if not is_duplicate:
                                tables.append(table_data)
                                if self.debug:
                                    print(f"  Found table: {table_data['rows']}x{table_data['cols']} at ({bbox[0]:.1f}, {bbox[1]:.1f})", file=sys.stderr)
                
            except Exception as e:
                if self.debug:
                    print(f"  Table extraction error: {e}", file=sys.stderr)
        
        return tables
    
    def assess_quality(self, result: Dict[str, Any]) -> float:
        """Assess the quality of extracted content"""
        total_score = 0.0
        score_weights = []
        
        # Assess table quality
        table_count = 0
        table_quality_sum = 0.0
        
        for page in result.get("pages", []):
            tables = page.get("tables", [])
            for table in tables:
                table_count += 1
                quality = self.assess_table_quality(table)
                table_quality_sum += quality
                self.quality_report["table_scores"].append({
                    "page": page["number"],
                    "rows": table.get("rows", 0),
                    "cols": table.get("cols", 0),
                    "quality": quality
                })
                
                if quality < 0.1:
                    self.quality_report["warnings"].append(
                        f"Page {page['number']}: Table with {table.get('rows', 0)}x{table.get('cols', 0)} "
                        f"is mostly empty (quality: {quality:.2f})"
                    )
        
        # Calculate table quality score
        if table_count > 0:
            avg_table_quality = table_quality_sum / table_count
            total_score += avg_table_quality * 0.5  # Tables are 50% of score
            score_weights.append(0.5)
            
            if avg_table_quality < 0.2:
                self.quality_report["errors"].append(
                    f"Table quality too low: {avg_table_quality:.2f} average"
                )
        
        # Assess text content quality
        text_quality = self.assess_text_quality(result)
        total_score += text_quality * 0.3  # Text is 30% of score
        score_weights.append(0.3)
        
        # Assess structural quality
        structure_quality = self.assess_structure_quality(result)
        total_score += structure_quality * 0.2  # Structure is 20% of score
        score_weights.append(0.2)
        
        # Normalize score
        if sum(score_weights) > 0:
            overall_score = total_score / sum(score_weights)
        else:
            overall_score = 0.0
            
        self.quality_report["overall_score"] = overall_score
        
        if self.debug:
            print(f"\n=== Quality Assessment ===", file=sys.stderr)
            print(f"Overall Score: {overall_score:.2f}", file=sys.stderr)
            print(f"Table Quality: {avg_table_quality:.2f}" if table_count > 0 else "No tables", file=sys.stderr)
            print(f"Text Quality: {text_quality:.2f}", file=sys.stderr)
            print(f"Structure Quality: {structure_quality:.2f}", file=sys.stderr)
            if self.quality_report["warnings"]:
                print(f"Warnings: {len(self.quality_report['warnings'])}", file=sys.stderr)
            if self.quality_report["errors"]:
                print(f"Errors: {len(self.quality_report['errors'])}", file=sys.stderr)
        
        return overall_score
    
    def assess_table_quality(self, table: Dict[str, Any]) -> float:
        """Assess quality of a single table"""
        data = table.get("data", [])
        if not data:
            return 0.0
        
        total_cells = 0
        non_empty_cells = 0
        
        for row in data:
            for cell in row:
                total_cells += 1
                if cell and str(cell).strip():
                    non_empty_cells += 1
        
        if total_cells == 0:
            return 0.0
            
        fill_ratio = non_empty_cells / total_cells
        
        # Penalize very small tables
        size_factor = min(1.0, (table.get("rows", 0) * table.get("cols", 0)) / 10.0)
        
        return fill_ratio * (0.7 + 0.3 * size_factor)
    
    def assess_text_quality(self, result: Dict[str, Any]) -> float:
        """Assess quality of text content"""
        total_chars = 0
        meaningful_elements = 0
        
        for page in result.get("pages", []):
            elements = page.get("elements", [])
            for element in elements:
                content = element.get("content", "").strip()
                if content:
                    total_chars += len(content)
                    # Count meaningful elements (not just single characters)
                    if len(content) > 2:
                        meaningful_elements += 1
        
        if total_chars == 0:
            return 0.0
        
        # Score based on content density
        pages_count = len(result.get("pages", []))
        if pages_count == 0:
            return 0.0
            
        avg_chars_per_page = total_chars / pages_count
        # Expect at least 100 characters per page for decent quality
        char_score = min(1.0, avg_chars_per_page / 100.0)
        
        # Factor in meaningful elements
        element_score = min(1.0, meaningful_elements / (pages_count * 5))
        
        return (char_score * 0.6 + element_score * 0.4)
    
    def assess_structure_quality(self, result: Dict[str, Any]) -> float:
        """Assess structural quality of the document"""
        heading_count = 0
        list_count = 0
        table_count = 0
        
        for page in result.get("pages", []):
            elements = page.get("elements", [])
            for element in elements:
                elem_type = element.get("type", "")
                if elem_type == "heading":
                    heading_count += 1
                elif elem_type == "list_item":
                    list_count += 1
            
            tables = page.get("tables", [])
            table_count += len(tables)
        
        # Score based on structural diversity
        structure_score = 0.0
        if heading_count > 0:
            structure_score += 0.4
        if list_count > 0:
            structure_score += 0.3
        if table_count > 0:
            structure_score += 0.3
            
        return structure_score


def main():
    parser = argparse.ArgumentParser(
        description="Extract text and structure from PDF with coordinate preservation"
    )
    parser.add_argument("pdf_file", help="Path to PDF file")
    parser.add_argument("--debug", "-d", action="store_true", 
                       help="Show debug information")
    parser.add_argument("--pretty", "-p", action="store_true",
                       help="Pretty print JSON output")
    parser.add_argument("--output", "-o", help="Output file (default: stdout)")
    parser.add_argument("--strict", "-s", action="store_true",
                       help="Strict quality mode - fail on low quality")
    parser.add_argument("--min-quality", "-q", type=float, default=0.2,
                       help="Minimum quality threshold (0.0-1.0, default: 0.2)")
    parser.add_argument("--ignore-quality", action="store_true",
                       help="Ignore quality checks and force extraction")
    
    args = parser.parse_args()
    
    # Check if file exists
    if not Path(args.pdf_file).exists():
        result = {
            "success": False,
            "error": f"File not found: {args.pdf_file}"
        }
        print(json.dumps(result))
        sys.exit(1)
    
    # Set quality parameters
    min_quality = 0.0 if args.ignore_quality else args.min_quality
    strict = False if args.ignore_quality else args.strict
    
    # Extract PDF content
    extractor = CoordinateBasedExtractor(
        args.pdf_file, 
        debug=args.debug,
        min_quality=min_quality,
        strict=strict
    )
    result = extractor.extract()
    
    # Determine exit code based on quality
    exit_code = 0
    if not result.get("success", True):
        if "quality too low" in result.get("error", "").lower():
            exit_code = 3  # Quality error
        else:
            exit_code = 1  # General error
    elif result.get("quality"):
        quality_score = result["quality"]["score"]
        if quality_score < min_quality and not args.ignore_quality:
            exit_code = 2  # Quality warning
    
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
    
    # Print quality summary to stderr if in debug mode
    if args.debug and result.get("quality"):
        report = result["quality"]["report"]
        if report.get("warnings"):
            print(f"\n⚠️  Warnings:", file=sys.stderr)
            for warning in report["warnings"]:
                print(f"  - {warning}", file=sys.stderr)
        if report.get("errors"):
            print(f"\n❌ Errors:", file=sys.stderr)
            for error in report["errors"]:
                print(f"  - {error}", file=sys.stderr)
    
    sys.exit(exit_code)


if __name__ == "__main__":
    main()