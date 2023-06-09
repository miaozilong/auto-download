package main

import (
	log "github.com/cihub/seelog"
	"github.com/dablelv/go-huge-util/zip"
	"net/http"
	os "os"
	"os/exec"
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
func download(w http.ResponseWriter, r *http.Request) {
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
	//设置响应头
	header := w.Header()
	header.Add("Content-Type", "application/octet-stream")
	header.Add("Content-Disposition", "attachment;filename="+projectName+".zip")
	var downloadByte []byte
	if downloadByte, err = os.ReadFile(zipFilePath); err != nil {
		log.Info("读取文件失败")
	}
	//写入到响应流中
	w.Write(downloadByte)
}
func main() {
	http.HandleFunc("/download", download)   //设置访问的路由
	err := http.ListenAndServe(":8081", nil) //设置监听的端口
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
