import sys
import os
try:
    from pdf2docx import Converter
except:
    os.system("pip install pdf2docx")
    from pdf2docx import Converter


params = sys.argv
pdf_file = params[1]
docx_file = params[2]

# convert pdf to docx
cv = Converter(pdf_file)
cv.convert(docx_file, start=0, end=None)
cv.close()
