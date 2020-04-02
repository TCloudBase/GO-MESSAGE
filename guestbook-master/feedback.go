package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/silenceper/goutils"
	"github.com/silenceper/wechat/tcb"
)

//FeedbackService service
type FeedbackService struct {
}

//NewFeedbackService new
func NewFeedbackService() *FeedbackService {
	return &FeedbackService{}
}

//Feedback 留言记录
type Feedback struct {
	Username   string `form:"username",json:"username"`
	Content    string `form:"content",json:"content"`
	FilePath   string `json:"filePath"`
	FileID     string `json:"fileId"`
	CreateTime string `json:"createTime"`
}

//List 文本列表
func (svc *FeedbackService) List(skip, limit int) ([]*Feedback, error) {
	query := fmt.Sprintf("db.collection(\"guestbook\").orderBy('createTime','desc').skip(%d).limit(%d).get()", skip, limit)
	data, err := getTcb().DatabaseQuery(getConfig().TcbEnv, query)
	if err != nil {
		return nil, err
	}
	feedbacks := make([]*Feedback, 0, len(data.Data))
	for _, v := range data.Data {
		feedbackItem := &Feedback{}
		err := json.Unmarshal([]byte(v), feedbackItem)
		if err != nil {
			return nil, err
		}
		feedbacks = append(feedbacks, feedbackItem)

	}
	//fmt.Println(data.Pager)
	return feedbacks, nil
}

//Count 统计记录数量
func (svc *FeedbackService) Count() (int64, error) {
	query := "db.collection(\"guestbook\").count()"
	res, err := getTcb().DatabaseCount(getConfig().TcbEnv, query)
	if err != nil {
		return 0, err
	}
	return res.Count, nil
}

//UploadFile 上传文件
func (svc *FeedbackService) UploadFile(path string, file io.Reader) (string, error) {
	//获取文件上传链接
	uploadRes, err := getTcb().UploadFile(getConfig().TcbEnv, path)
	if err != nil {
		return "", err
	}

	data := make(map[string]io.Reader)
	data["key"] = strings.NewReader(path)
	data["Signature"] = strings.NewReader(uploadRes.Authorization)
	data["x-cos-security-token"] = strings.NewReader(uploadRes.Token)
	data["x-cos-meta-fileid"] = strings.NewReader(uploadRes.CosFileID)
	data["file"] = file

	//上传文件
	_, err = goutils.PostFormWithFile(&http.Client{}, uploadRes.URL, data)
	return uploadRes.FileID, err
}

//DownloadFile 获取下载链接
func (svc *FeedbackService) DownloadFile(id string) (string, error) {
	files := []*tcb.DownloadFile{&tcb.DownloadFile{
		FileID: id,
		MaxAge: 100,
	}}
	res, err := getTcb().BatchDownloadFile(getConfig().TcbEnv, files)
	if err != nil {
		return "", err
	}
	if len(res.FileList) >= 1 {
		return res.FileList[0].DownloadURL, nil
	}
	return "", nil
}

//Save 保存内容
func (svc *FeedbackService) Save(feedback *Feedback) error {
	if feedback.Username == "" || feedback.Content == "" {
		return fmt.Errorf("用户名或留言内容不能为空")
	}
	//content 调用云函数过滤
	var err error
	feedback.Content, err = svc.FilterText(feedback.Content)
	if err != nil {
		return err
	}
	query := `db.collection(\"%s\").add({
      data: [{
        username: \"%s\",
        content: \"%s\",
		filePath: \"%s\",
		fileId: \"%s\",
        createTime: \"%s\",
      }]
      })`
	feedback.CreateTime = time.Now().Format("2006-01-02 15:04:05")
	query = fmt.Sprintf(query, "guestbook", feedback.Username, feedback.Content, feedback.FilePath, feedback.FileID, feedback.CreateTime)
	_, err = getTcb().DatabaseAdd(getConfig().TcbEnv, query)
	if err != nil {
		return err
	}
	return nil
}

//FilterRes 过滤文件的结果
type FilterRes struct {
	Text string `json:"text"`
}

//FilterText 调用云函数过滤文本
func (svc *FeedbackService) FilterText(text string) (string, error) {
	res, err := getTcb().InvokeCloudFunction(getConfig().TcbEnv, "filterText", fmt.Sprintf(`{"text":"%s"}`, text))
	//返回的是json
	filterRes := &FilterRes{}
	err = json.Unmarshal([]byte(res.RespData), filterRes)
	if err != nil {
		return "", nil
	}

	return filterRes.Text, nil
}
