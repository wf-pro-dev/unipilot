#!/usr/bin/env python3
import docx2txt
import sys

if len(sys.argv) != 2:
    print("Usage: python3 extract_docx.py <docx-file>")
    sys.exit(1)

filename = sys.argv[1]
try:
    text = docx2txt.process(filename)
    print(text)
except Exception as e:
    print(f"Error: {e}")
    sys.exit(1)
