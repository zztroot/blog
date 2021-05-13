import comtypes.client
import sys

#文件要使用绝对路径

def convertDocxToPDF(infile,outfile):
    wdFormatPDF = 17
    word = comtypes.client.CreateObject('Word.Application')
    doc = word.Documents.Open(infile)
    doc.SaveAs(outfile, FileFormat=wdFormatPDF)
    doc.Close()
    word.Quit()


if __name__ == '__main__':
    params = sys.argv
    word_file = params[1]
    pdf_file = params[2]
    convertDocxToPDF(word_file, pdf_file)