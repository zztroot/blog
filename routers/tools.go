package routers

import "zzt_blog/middles"

func (r *RouterStruct) ToolsRouters(){
	//工具相关路由
	tools := r.Router.Group("/api/tools")
	tools.POST("/uploadFile", r.ToolsService.PdfToWord)
	tools.POST("/uploadFileWordToPdf", r.ToolsService.WordToPdf)
	tools.POST("/uploadFilePdfToImage", r.ToolsService.PdfToImage)
	//HTML转PDF
	tools.POST("/uploadFileHtmlToPdf", r.ToolsService.HtmlToPdf)
	tools.GET("/downloadFile", r.ToolsService.DownloadFile)
	tools.GET("/getImageZipSize", r.ToolsService.GetFileZipSize)
	//中间件
	tools.Use(middles.FilterSql())
	{
		//md5加密
		tools.POST("/md5Encode", r.ToolsService.Md5Encode)
		//md5解密
		tools.POST("/md5Decode", r.ToolsService.Md5Decode)

		//base64加密
		tools.POST("/base64Encode", r.ToolsService.Base64Encode)
		//base64解密
		tools.POST("/base64Decode", r.ToolsService.Base64Decode)

		//URL加密
		tools.POST("/urlEncode", r.ToolsService.URLEncode)
		//URL解密
		tools.POST("/urlDecode", r.ToolsService.URLDecode)

		//中文转unicode
		tools.POST("/chineseToUnicode", r.ToolsService.ChineseToUnicode)
		//unicode转中文
		tools.POST("/unicodeToChinese", r.ToolsService.UnicodeToChinese)
	}
}
