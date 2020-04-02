package main

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	//定义路由信息
	r.LoadHTMLGlob("template/*")

	//渲染首页
	r.GET("/", index)
	//提交留言
	r.POST("/feedback", feedback)
	//下载附件
	r.GET("/file", downloadFile)
	log.Fatal(r.Run(":8080")) // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}

//首页
func index(c *gin.Context) {
	page := c.Query("page")
	//获取记录数量
	feedbackService := NewFeedbackService()
	count, err := feedbackService.Count()
	if err != nil {
		showError(c, err)
		return
	}
	limit := 10
	totalPage := math.Ceil(float64(count) / float64(limit))
	totalPageInt := int(totalPage)
	pageInt, _ := strconv.Atoi(page)
	if pageInt > totalPageInt {
		pageInt = totalPageInt
	}
	if pageInt < 1 {
		pageInt = 1
	}

	//展示留言列表
	skip := (pageInt - 1) * limit
	list, err := feedbackService.List(skip, limit)
	if err != nil {
		showError(c, err)
		return
	}

	c.HTML(http.StatusOK, "index.html", gin.H{
		"title":     "留言板",
		"list":      list,
		"prevPage":  pageInt - 1,
		"nextPage":  pageInt + 1,
		"page":      pageInt,
		"totalPage": totalPageInt,
	})
}

//留言
func feedback(c *gin.Context) {
	//1、接收提交参数
	feedback := &Feedback{}
	err := c.Bind(feedback)
	if err != nil {
		showError(c, err)
		return
	}
	//2、文件上传
	fileHeader, err := c.FormFile("file")
	if err != nil && err != http.ErrMissingFile {
		showError(c, err)
		return
	}
	feedbackService := NewFeedbackService()
	if fileHeader != nil {
		path := fmt.Sprintf("guestbook/%s", fileHeader.Filename)
		file, err := fileHeader.Open()
		if err != nil {
			showError(c, err)
			return
		}

		fileID, err := feedbackService.UploadFile(path, file)
		if err != nil {
			showError(c, err)
			return
		}
		feedback.FilePath = fileHeader.Filename
		feedback.FileID = fileID
	}

	//保存内容
	err = feedbackService.Save(feedback)
	if err != nil {
		showError(c, err)
		return
	}

	c.Redirect(http.StatusMovedPermanently, "/")
}

//附件下载
func downloadFile(c *gin.Context) {
	fileID := c.Query("id")
	if fileID == "" {
		showError(c, fmt.Errorf("fileID为空"))
		return
	}
	downLoadURL, err := NewFeedbackService().DownloadFile(fileID)
	if err != nil {
		showError(c, err)
		return
	}
	c.Redirect(http.StatusMovedPermanently, downLoadURL)
}

func showError(c *gin.Context, err error) {
	c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("%s", err)})
}
