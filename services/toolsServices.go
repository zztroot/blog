package services

import (
	"crypto/md5"
	"archive/zip"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	logger "github.com/zztroot/zztlog"
	"io"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
	. "zzt_blog/common"
	"zzt_blog/databases"
	"zzt_blog/models"
)

type ToolsService struct {
	db               databases.MysqlConn
	downloadFileList map[string][]string
	isOn             bool
}

func (t *ToolsService) GetFileZipSize(c *gin.Context) {
	list, ok := t.downloadFileList[c.Request.Host]
	if !ok {
		logger.Error("文件不存在")
		c.JSON(200, gin.H{
			"error":"请先上传文件转换",
		})
		return
	}
	open, err := os.Stat(list[0])
	if err != nil {
		logger.Error(err)
	}
	c.JSON(200, gin.H{
		"file_size":open.Size(),
		"error":"",
	})
}

//清理downloadFileList map内存数据
func (t *ToolsService) clearMemory() {
	for {
		for k, v := range t.downloadFileList {
			nowTime := time.Now().Add(-time.Minute * 10).Format("2006-01-02 15:04:05")
			if v[1] < nowTime {
				delete(t.downloadFileList, k)
			}
		}
		time.Sleep(time.Minute * 10)
	}
}

//下载PDF转WORD文件
func (t *ToolsService) DownloadFile(c *gin.Context) {
	file, ok := t.downloadFileList[c.Request.Host]
	if !ok {
		c.JSON(200, gin.H{
			"error": "请先转换文件",
		})
		return
	}
	logger.Debug("1231---", file)
	readFile, err := ioutil.ReadFile(file[0])
	if err != nil {
		c.JSON(200, gin.H{
			"error": "请先转换文件",
		})
		delete(t.downloadFileList, c.Request.Host)
		logger.Error(err)
		return
	}
	c.Header("response-type", "blob")
	c.Header("content-disposition", `attachment; filename=`+file[0])
	c.Data(200, "application/vnd.openxmlformats-officedocument.wordprocessingml.document", readFile)
	//temp := strings.Split(file, "/")
	//fileName := temp[len(temp)-1]
	//c.JSON(200, gin.H{
	//	"file_name":fileName,
	//	"file_path":file,
	//	"error":"",
	//})
}

//PDF转WORD
func (t *ToolsService) PdfToWord(c *gin.Context) {
	var (
		errList  string
		filePath []string
	)
	if !t.isOn {
		t.isOn = true
		go t.clearMemory()
	}
	value, ok := t.downloadFileList[c.Request.Host]
	if ok {
		go func(value string) {
			if err := os.Remove(value); err != nil {
				logger.Error(err)
			}
			t.delFile(value)
		}(value[0])
		delete(t.downloadFileList, c.Request.Host)
	}
	file, err := c.MultipartForm()
	if err != nil {
		logger.Error(err)
		return
	}
	files := file.File["fileBlob"]
	for _, f := range files {
		if path.Ext(f.Filename) == ".pdf" || path.Ext(f.Filename) == ".PDF" {
			//判断文件是否存在
			open, err := os.Open("media/" + f.Filename)
			if err == nil {
				errList = "文件名已经存在:" + f.Filename
				open.Close()
				break
			}
			tempTime := time.Now().Nanosecond()
			newPath := fmt.Sprintf("media/%d-%s", tempTime, f.Filename)
			create, err := os.Create(newPath)
			if err != nil {
				logger.Error(err)
			}
			data, err := f.Open()
			if err != nil {
				logger.Error(err)
			}
			_, err = io.Copy(create, data)
			if err != nil {
				logger.Error(err)
			}
			create.Close()
			//转换
			toPath := fmt.Sprintf("media/%d-%s.docx", tempTime, strings.TrimSuffix(f.Filename, ".pdf"))
			if err := t.handlerPdfToWord(newPath, toPath); err != nil {
				errList = "转换失败"
				break
			}
			filePath = append(filePath, toPath)
			t.downloadFileList = map[string][]string{
				c.Request.Host: {toPath, time.Now().Format("2006-01-02 15:04:05")},
			}
		} else {
			if errList == "" {
				errList = f.Filename
			} else {
				errList = errList + "," + f.Filename
			}
		}
	}
	if errList != "" && errList != "转换失败"{
		errList = errList + "--(错误,请上传pdf格式文件)"
	}
	c.JSON(200, gin.H{
		"error":                   errList,
		"initialPreview":          []string{},
		"initialPreviewConfig":    []string{},
		"initialPreviewThumbTags": []string{},
		"append":                  true,
		"filePath":                filePath,
		//"msgUploadEnd":"合成完成",
	})
}

//删除转换后的文件
func (t *ToolsService) delFile(value string) {
	switch path.Ext(value) {
	case ".pdf":
		//pdf转word"
		value = strings.ReplaceAll(value, ".pdf", ".docx")
		_, err := os.Stat(value)
		if err != nil {
			value = strings.ReplaceAll(value, ".docx", ".html")
		}
	case ".docx":
		//word转pdf"
		value = strings.ReplaceAll(value, ".docx", ".pdf")
	case ".zip":
		split := strings.Split(value, ".")
		value = fmt.Sprintf("%s.pdf", split[0])
	}
	if err := os.Remove(value); err != nil {
		logger.Error(err)
	}
}

//WORD转PDF
func (t *ToolsService) WordToPdf(c *gin.Context) {
	var (
		errList  string
		filePath []string
	)
	if !t.isOn {
		t.isOn = true
		go t.clearMemory()
	}
	value, ok := t.downloadFileList[c.Request.Host]
	if ok {
		go func(value string) {
			if err := os.Remove(value); err != nil {
				logger.Error(err)
			}
			t.delFile(value)
		}(value[0])
		delete(t.downloadFileList, c.Request.Host)
	}
	file, err := c.MultipartForm()
	if err != nil {
		logger.Error(err)
		return
	}
	files := file.File["fileBlob"]
	for _, f := range files {
		if path.Ext(f.Filename) == ".docx" {
			//判断文件是否存在
			open, err := os.Open("media/" + f.Filename)
			if err == nil {
				errList = "文件名已经存在:" + f.Filename
				open.Close()
				break
			}
			tempTime := time.Now().Nanosecond()
			//获取当前绝对路径
			getPath, err := os.Getwd()
			if err != nil {
				logger.Error(err)
			}

			newPath := fmt.Sprintf("%s%cmedia%c%d-%s", getPath, os.PathSeparator, os.PathSeparator, tempTime, f.Filename)
			create, err := os.Create(newPath)
			if err != nil {
				logger.Error(err)
			}
			data, err := f.Open()
			if err != nil {
				logger.Error(err)
			}
			_, err = io.Copy(create, data)
			if err != nil {
				logger.Error(err)
			}
			create.Close()
			//转换
			toPath := fmt.Sprintf("%s%cmedia%c%d-%s.pdf", getPath, os.PathSeparator, os.PathSeparator, tempTime, strings.TrimSuffix(f.Filename, ".docx"))
			tempPath := fmt.Sprintf("%s%cmedia%c", getPath, os.PathSeparator, os.PathSeparator)
			if err := t.handlerWordToPdf(newPath, tempPath); err != nil {
				errList = "转换失败"
				break
			}
			filePath = append(filePath, toPath)
			t.downloadFileList = map[string][]string{
				c.Request.Host: {toPath, time.Now().Format("2006-01-02 15:04:05")},
			}
		} else {
			if errList == "" {
				errList = f.Filename
			} else {
				errList = errList + "," + f.Filename
			}
		}
	}
	if errList != "" && errList != "转换失败"{
		errList = errList + "--(错误,请上传docx格式文件)"
	}
	c.JSON(200, gin.H{
		"error":                   errList,
		"initialPreview":          []string{},
		"initialPreviewConfig":    []string{},
		"initialPreviewThumbTags": []string{},
		"append":                  true,
		"filePath":                filePath,
		//"msgUploadEnd":"合成完成",
	})
}

//PDF转图片
func (t *ToolsService) PdfToImage(c *gin.Context) {
	var (
		errList  string
		filePath []string
	)
	if !t.isOn {
		t.isOn = true
		go t.clearMemory()
	}
	value, ok := t.downloadFileList[c.Request.Host]
	if ok {
		go func(value string) {
			if err := os.Remove(value); err != nil {
				logger.Error(err)
			}
			t.delFile(value)
		}(value[0])
		delete(t.downloadFileList, c.Request.Host)
	}
	file, err := c.MultipartForm()
	if err != nil {
		logger.Error(err)
		return
	}
	files := file.File["fileBlob"]
	for _, f := range files {
		if path.Ext(f.Filename) == ".pdf" || path.Ext(f.Filename) == ".PDF" {
			//判断文件是否存在
			open, err := os.Open("media/" + f.Filename)
			if err == nil {
				errList = "文件名已经存在:" + f.Filename
				open.Close()
				break
			}
			tempTime := time.Now().Nanosecond()
			newPath := fmt.Sprintf("media/%d-%s", tempTime, f.Filename)
			create, err := os.Create(newPath)
			if err != nil {
				logger.Error(err)
			}
			data, err := f.Open()
			if err != nil {
				logger.Error(err)
			}
			_, err = io.Copy(create, data)
			if err != nil {
				logger.Error(err)
			}
			create.Close()
			//转换
			toPath := fmt.Sprintf("media%c%d-%s.png",os.PathSeparator, tempTime, strings.TrimSuffix(f.Filename, ".pdf"))
			if err := t.handlerPdfToImage(newPath, toPath); err != nil {
				errList = "转换失败"
				break
			}
			//创建文件夹
			dirPath := fmt.Sprintf("media%c%d-%s", os.PathSeparator, tempTime, strings.TrimSuffix(f.Filename, ".pdf"))
			_ = os.Mkdir(dirPath, os.ModePerm)

			toZip := t.handlerImageToZip(dirPath, fmt.Sprintf("%d-%s", tempTime, strings.TrimSuffix(f.Filename, ".pdf")))

			filePath = append(filePath, toZip)
			t.downloadFileList = map[string][]string{
				c.Request.Host: {toZip, time.Now().Format("2006-01-02 15:04:05")},
			}
		} else {
			if errList == "" {
				errList = f.Filename
			} else {
				errList = errList + "," + f.Filename
			}
		}
	}
	if errList != "" && errList != "转换失败"{
		errList = errList + "--(错误,请上传pdf格式文件)"
	}
	c.JSON(200, gin.H{
		"error":                   errList,
		"initialPreview":          []string{},
		"initialPreviewConfig":    []string{},
		"initialPreviewThumbTags": []string{},
		"append":                  true,
		"filePath":                filePath,
		//"msgUploadEnd":"合成完成",
	})
}

//HTML转pdf
func (t *ToolsService) HtmlToPdf(c *gin.Context) {
	var (
		errList  string
		filePath []string
	)
	if !t.isOn {
		t.isOn = true
		go t.clearMemory()
	}
	value, ok := t.downloadFileList[c.Request.Host]
	if ok {
		go func(value string) {
			if err := os.Remove(value); err != nil {
				logger.Error(err)
			}
			t.delFile(value)
		}(value[0])
		delete(t.downloadFileList, c.Request.Host)
	}
	file, err := c.MultipartForm()
	if err != nil {
		logger.Error(err)
		return
	}
	files := file.File["fileBlob"]
	for _, f := range files {
		if path.Ext(f.Filename) == ".html" {
			//判断文件是否存在
			open, err := os.Open("media/" + f.Filename)
			if err == nil {
				errList = "文件名已经存在:" + f.Filename
				open.Close()
				break
			}
			tempTime := time.Now().Nanosecond()
			//获取当前绝对路径
			//getPath, err := os.Getwd()
			//if err != nil {
			//	logger.Error(err)
			//}

			newPath := fmt.Sprintf("media%c%d-%s",os.PathSeparator, tempTime, f.Filename)
			create, err := os.Create(newPath)
			if err != nil {
				logger.Error(err)
			}
			data, err := f.Open()
			if err != nil {
				logger.Error(err)
			}
			_, err = io.Copy(create, data)
			if err != nil {
				logger.Error(err)
			}
			create.Close()
			//转换
			toPath := fmt.Sprintf("media%c%d-%s.pdf", os.PathSeparator, tempTime, strings.TrimSuffix(f.Filename, ".html"))
			if err := t.handlerHtmlToPdf(newPath, toPath); err != nil {
				errList = "转换失败"
				break
			}
			filePath = append(filePath, toPath)
			t.downloadFileList = map[string][]string{
				c.Request.Host: {toPath, time.Now().Format("2006-01-02 15:04:05")},
			}
		} else {
			if errList == "" {
				errList = f.Filename
			} else {
				errList = errList + "," + f.Filename
			}
		}
	}
	if errList != "" && errList != "转换失败" {
		errList = errList + "--(错误,请上传html格式文件)"
	}
	c.JSON(200, gin.H{
		"error":                   errList,
		"initialPreview":          []string{},
		"initialPreviewConfig":    []string{},
		"initialPreviewThumbTags": []string{},
		"append":                  true,
		"filePath":                filePath,
		//"msgUploadEnd":"合成完成",
	})
}

//图片打包zip
func (t *ToolsService) handlerImageToZip(dir string, imgName string) string{
	files, err := ioutil.ReadDir("media")
	if err != nil {
		logger.Error(err)
	}
	for _, v := range files {
		if strings.Contains(v.Name(), imgName) && path.Ext(v.Name()) == ".png"{
			//var command *exec.Cmd
			if runtime.GOOS == "windows" {
				//command = exec.Command("move", fmt.Sprintf("media%c%s", os.PathSeparator, v.Name()), fmt.Sprintf("%s%c%s", dir, os.PathSeparator, v.Name()))
				if err := os.Rename(fmt.Sprintf("media%c%s", os.PathSeparator, v.Name()), fmt.Sprintf("%s%c%s", dir, os.PathSeparator, v.Name())); err != nil {
					logger.Error(err)
				}
				//err := os.Remove(fmt.Sprintf("media%c%s", os.PathSeparator, v.Name()))
			}else {
				command := exec.Command("mv", fmt.Sprintf("media%c%s", os.PathSeparator, v.Name()), fmt.Sprintf("%s%c%s", dir, os.PathSeparator, v.Name()))
				if err := command.Run(); err != nil {
					logger.Error(err)
				}
			}

		}
	}
	zipPath := fmt.Sprintf("media%c%s.zip", os.PathSeparator, imgName)
	ZipDir(dir, zipPath)
	go func(dir string) {
		if err := os.RemoveAll(dir); err != nil {
			logger.Error(err)
		}
	}(dir)
	return zipPath
}

func ZipDir(dir, zipFile string) {

	fz, err := os.Create(zipFile)
	if err != nil {
		log.Printf("Create zip file failed: %s\n", err.Error())
	}
	defer fz.Close()

	w := zip.NewWriter(fz)
	defer w.Close()

	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			fDest, err := w.Create(path[len(dir)+1:])
			if err != nil {
				log.Printf("Create failed: %s\n", err.Error())
				return nil
			}
			fSrc, err := os.Open(path)
			if err != nil {
				log.Printf("Open failed: %s\n", err.Error())
				return nil
			}
			defer fSrc.Close()
			_, err = io.Copy(fDest, fSrc)
			if err != nil {
				log.Printf("Copy failed: %s\n", err.Error())
				return nil
			}
		}
		return nil
	})
}


func (t *ToolsService) handlerPdfToWord(p string, to string) error{
	command := exec.Command("python3", "script/python/pdfToWord.py", p, to)
	if err := command.Run(); err != nil {
		logger.Error(err)
		return err
	}
	return nil
}

func (t *ToolsService) handlerWordToPdf(p string, to string) error{
	//soffice --headless --invisible --convert-to pdf:writer_pdf_Export
	args := []string{
		"--headless",
		"--invisible",
		"--convert-to",
		"pdf:writer_pdf_Export",
		p,
		"--outdir",
		to,
	}
	command := exec.Command("soffice", args...)
	if err := command.Run(); err != nil {
		logger.Error(err)
		return err
	}
	return nil
}

func (t *ToolsService) handlerHtmlToPdf(p string, to string) error{
	args := []string{
		p,
		to,
	}
	command := exec.Command("wkhtmltopdf", args...)
	if err := command.Run(); err != nil {
		logger.Error(err)
		return err
	}
	return nil
}

func (t *ToolsService) handlerPdfToImage(p string, to string) error {
	var command *exec.Cmd
	if runtime.GOOS != "windows" {
		command = exec.Command("tools/pdftopng", p, to)

	}else {
		command = exec.Command("tools/pdftopng.exe", p, to)
	}
	if err := command.Run(); err != nil {
		logger.Error(err)
		return err
	}
	return nil
}

//中文转unicode
func (t *ToolsService) ChineseToUnicode(c *gin.Context) {
	temp, _ := c.Get("data")
	data := temp.(map[string]interface{})
	input := data["input"].(string)
	textQuoted := strconv.QuoteToASCII(input)
	results := textQuoted[1 : len(textQuoted)-1]
	m := make(map[string]interface{})
	m["data"] = results
	bs, _ := json.Marshal(m)
	Reply(c, SlotReply{Data: bs})
}

//unicode转中文
func (t *ToolsService) UnicodeToChinese(c *gin.Context) {
	temp, _ := c.Get("data")
	data := temp.(map[string]interface{})
	input := data["input"].(string)
	sUnicode := strings.Split(input, "\\u")
	var results string
	for _, v := range sUnicode {
		if len(v) < 1 {
			continue
		}
		temp, err := strconv.ParseInt(v, 16, 32)
		if err != nil {
			logger.Error(err)
			return
		}
		results += fmt.Sprintf("%c", temp)
	}
	m := make(map[string]interface{})
	m["data"] = results
	bs, _ := json.Marshal(m)
	Reply(c, SlotReply{Data: bs})
}

//URL加密
func (t *ToolsService) URLEncode(c *gin.Context) {
	temp, _ := c.Get("data")
	data := temp.(map[string]interface{})
	input := data["input"].(string)

	if !strings.Contains(input, "http") && !strings.Contains(input, "https") {
		Reply(c, SlotReply{Error: "请输入正确的URL"})
		return
	}
	results := url.QueryEscape(input)
	m := make(map[string]interface{})
	m["data"] = results
	bs, _ := json.Marshal(m)
	Reply(c, SlotReply{Data: bs})
}

//URL解密
func (t *ToolsService) URLDecode(c *gin.Context) {
	temp, _ := c.Get("data")
	data := temp.(map[string]interface{})
	input := data["input"].(string)

	results, err := url.QueryUnescape(input)
	if err != nil {
		logger.Error(err)
		Reply(c, SlotReply{Error: "解密失败"})
		return
	}
	m := make(map[string]interface{})
	m["data"] = results
	bs, _ := json.Marshal(m)
	Reply(c, SlotReply{Data: bs})
}

//Base64加密
func (t *ToolsService) Base64Encode(c *gin.Context) {
	temp, _ := c.Get("data")
	data := temp.(map[string]interface{})
	input := data["input"].(string)

	results := base64.StdEncoding.EncodeToString([]byte(input))
	m := make(map[string]interface{})
	m["data"] = results
	bs, _ := json.Marshal(m)
	Reply(c, SlotReply{Data: bs})
}

//Base64解密
func (t *ToolsService) Base64Decode(c *gin.Context) {
	temp, _ := c.Get("data")
	data := temp.(map[string]interface{})
	input := data["input"].(string)

	decoded, err := base64.StdEncoding.DecodeString(input)
	if err != nil {
		logger.Error(err)
		Reply(c, SlotReply{Error: "解密失败"})
		return
	}

	m := make(map[string]interface{})
	m["data"] = string(decoded)
	bs, _ := json.Marshal(m)
	Reply(c, SlotReply{Data: bs})
}

//MD5解密
func (t *ToolsService) Md5Decode(c *gin.Context) {
	temp, _ := c.Get("data")
	data := temp.(map[string]interface{})
	input := data["input"].(string)
	db := t.db.ConnectMysql().(*gorm.DB)

	var results string

	//查询数据库
	md5Str := models.Md5Info{}
	if err := db.Where("md5_value = ?", input).Find(&md5Str).Error; err != nil {
		logger.Error(err)
		return
	}
	results = md5Str.ResultValue
	logger.Info(results)
	if results == "" {
		Reply(c, SlotReply{Error: "解密失败"})
	} else {
		m := make(map[string]interface{})
		m["data"] = results
		bs, _ := json.Marshal(m)
		Reply(c, SlotReply{Data: bs})
	}
}

//MD5加密
func (t *ToolsService) Md5Encode(c *gin.Context) {
	temp, _ := c.Get("data")
	data := temp.(map[string]interface{})
	input := data["input"].(string)
	db := t.db.ConnectMysql().(*gorm.DB)

	output32 := Get32MD5Encode(input)
	output16 := Get16MD5Encode(input)

	go func() {
		//查询数据库
		md5Str := models.Md5Info{}
		if err := db.Where("md5_value = ? or md5_value = ?", output32, output16).Find(&md5Str).Error; err == nil {
			return
		} else {
			if err == gorm.ErrRecordNotFound {
				//保存32位
				md5Str1 := models.Md5Info{}
				md5Str1.CreatedAt = GetTime()
				md5Str1.Md5Value = output32
				md5Str1.ResultValue = input
				if err := db.Save(&md5Str1).Error; err != nil {
					logger.Error(err)
					return
				}

				//保存16位
				md5Str2 := models.Md5Info{}
				md5Str2.CreatedAt = GetTime()
				md5Str2.Md5Value = output16
				md5Str2.ResultValue = input
				if err := db.Save(&md5Str2).Error; err != nil {
					logger.Error(err)
					return
				}
			}
		}
	}()

	//这里必须换行，前端显示才能换行
	results := fmt.Sprintf(`16位=%s
32位=%s`, output16, output32)
	m := make(map[string]interface{})
	m["data"] = results
	bs, _ := json.Marshal(m)
	Reply(c, SlotReply{Data: bs})
}

//返回一个32位md5加密后的字符串
func Get32MD5Encode(data string) string {
	h := md5.New()
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

//返回一个16位md5加密后的字符串
func Get16MD5Encode(data string) string {
	return Get32MD5Encode(data)[8:24]
}
