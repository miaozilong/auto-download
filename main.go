package main

import (
	"encoding/json"
	log "github.com/cihub/seelog"
	"github.com/dablelv/go-huge-util/zip"
	"net/http"
	os "os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type Body struct {
	Url string
}

func init() {
	logger, err := log.LoggerFromConfigAsFile("seelog.xml")
	if err != nil {
		return
	}
	_ = log.ReplaceLogger(logger)
	log.Info("项目启动")
}
func downloadReq(w http.ResponseWriter, r *http.Request) {
	url := r.FormValue("url")
	log.Debug("接收到下载请求", url)
	projectName := getProjectNameFromUrl(url)
	nowStr := time.Now().Format("20060102150405")
	cloneTimePath := "/download/" + nowStr
	clonePath := cloneTimePath + "/" + projectName
	cmd := exec.Command("git", "clone", "--depth", "1", url, clonePath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Info(err)
	}
	s := string(output)
	log.Info("下载成功，内容为:\n" + s)
	zipFilePath := cloneTimePath + "/" + projectName + ".zip"
	_ = zip.Zip(zipFilePath, clonePath)
	log.Info("压缩成功")
	// 返回 filePath
	resp := make(map[string]string)
	resp["filePath"] = zipFilePath
	jsonResp, _ := json.Marshal(resp)
	// 以下两行顺序不能颠倒
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResp)
}
func downloadFile(w http.ResponseWriter, r *http.Request) {
	var err error
	filePath := r.FormValue("filePath")
	fileName := getFileNameFromFilePath(filePath)
	//设置响应头
	header := w.Header()
	header.Add("Content-Type", "application/octet-stream")
	header.Add("Content-Disposition", "attachment;filename="+fileName)
	var downloadByte []byte
	if downloadByte, err = os.ReadFile(filePath); err != nil {
		log.Info("读取文件失败")
	}
	//写入到响应流中
	w.Write(downloadByte)
}
func main() {
	go func() {
		for {
			go AutoDelete("/download")
			time.Sleep(time.Hour)
		}
	}()
	http.HandleFunc("/downloadReq", downloadReq)   //设置访问的路由
	http.HandleFunc("/downloadFile", downloadFile) //设置访问的路由
	err := http.ListenAndServe(":8081", nil)       //设置监听的端口
	if err != nil {
		log.Info("ListenAndServe: ", err)
	}
}

func getProjectNameFromUrl(url string) string {
	var ret string
	split := strings.Split(url, "/")
	s2 := split[len(split)-1]
	i := strings.Split(s2, ".")
	ret = i[0]
	return ret
}

func getFileNameFromFilePath(path string) string {
	var ret = ""
	split := strings.Split(path, "/")
	if len(split) > 0 {
		ret = split[len(split)-1]
	}
	return ret
}

// 自动删除48小时之前的文件
func AutoDelete(url string) {
	var files []os.FileInfo
	var filesPath []string

	root := url
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		files = append(files, info)
		filesPath = append(filesPath, path)
		log.Debug("添加了路径:" + path)
		return nil
	})
	if err != nil {
		log.Debug(err)
		panic(err)
	}
	if files != nil {
		log.Debug("文件列表:")
		log.Debug(files)
		for index, file := range files {
			if file != nil && !file.IsDir() {
				fileTime := file.ModTime()
				log.Debug("文件时间：" + fileTime.Format("2006-01-02 15:04:05"))
				if time.Now().Sub(fileTime) > time.Hour*48 {
					filePath := filesPath[index]
					os.Remove(filePath)
					log.Debug("删除了" + filePath + "，文件的修改日期是" + fileTime.Format("2006-01-02 15:04:05"))
				}
			}
		}
	}
}
