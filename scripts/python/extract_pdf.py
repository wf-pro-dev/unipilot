#!/usr/bin/env python3
from pypdf import PdfReader
import sys

if len(sys.argv) != 2:
    print("Usage: python3 extract_pdf.py <docx-file>")
    sys.exit(1)

filename = sys.argv[1]
try:
    reader = PdfReader(filename)
    text = ""
    for page in reader.pages:
        text += page.extract_text() + "\n"
    print(text)
except Exception as e:
    print(f"Error: {e}")
    sys.exit(1)


