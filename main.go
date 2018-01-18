package main

import (
	"net/http"

	"log"

	"encoding/json"
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"io"
	"io/ioutil"
	"regexp"
)

func main() {
	http.HandleFunc("/ueditor", GetUeditorConfig) //设置访问的路由
	err := http.ListenAndServe(":9090", nil)      //设置监听的端口
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

const (
	OssEndpoint = "OssEndpoint"
	OssKey      = "OssKey"
	OssSecret   = "OssSecret"
	BucketName  = "BucketName"
	OssUrl      = "OssUrl"
)

func GetUeditorConfig(w http.ResponseWriter, r *http.Request) {

	action := r.FormValue("action")

	switch action {
	//自动读入配置文件，只要初始化UEditor即会发生
	case "config":

		jsonByte, _ := ioutil.ReadFile("config/ueditor.json")
		re, _ := regexp.Compile("\\/\\*[\\S\\s]+?\\*\\/")
		jsonByte = re.ReplaceAll(jsonByte, []byte(""))

		w.Write(jsonByte)
		w.Header().Add("Content-Type", "application/json; charset=utf-8")

	case "uploadimage":
		{
			upload(w, r)

		}

	case "uploadfile":
		{
			upload(w, r)
		}

	}
}

func upToOss(filename string, reader io.Reader) (bool, error) {
	client, err := oss.New(OssEndpoint, OssKey, OssSecret)
	if err != nil {

		return false, err
	}

	bucket, err := client.Bucket(BucketName)
	if err != nil {

		return false, err
	}

	//objectKey string, reader io.Reader, options ...Option
	err = bucket.PutObject(filename, reader)
	if err != nil {

		return false, err
	}
	return true, nil
}

func upload(w http.ResponseWriter, r *http.Request) {
	//获取上传的文件，直接可以获取表单名称对应的文件名，不用另外提取
	file, header, err := r.FormFile("upfile")
	if err != nil {
		data, _ := json.Marshal(map[string]string{
			"state": fmt.Sprintf("获取文件错误: %s", err.Error()),
		})
		w.Write(data)
		return
	}

	//可以获取随机数
	name := header.Filename

	isUp, err := upToOss(name, file)

	if isUp {
		data, _ := json.Marshal(map[string]string{
			"url":      fmt.Sprintf(OssUrl + name), //保存后的文件路径
			"title":    "",                         //文件描述，对图片来说在前端会添加到title属性上
			"original": header.Filename,            //原始文件名
			"state":    "SUCCESS",                  //上传状态，成功时返回SUCCESS,其他任何值将原样返回至图片上传框中
		})
		w.Write(data)
		return
	} else if err != nil {
		data, _ := json.Marshal(map[string]string{
			"state": fmt.Sprintf("上传失败: %s", err.Error()),
		})
		w.WriteHeader(400)
		w.Write(data)
		return
	}
}
