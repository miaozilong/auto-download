package main

import (
	"encoding/json"
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
func sayhelloName(w http.ResponseWriter, r *http.Request) {
	var body Body
	var downloadByte []byte
	json.NewDecoder(r.Body).Decode(&body)
	url := body.Url
	projectName := getProjectNameFromUrl(url)
	nowStr := time.Now().Format("20060102150405")
	cloneTimePath := "/download/" + nowStr
	clonePath := cloneTimePath + "/" + projectName
	cmd := exec.Command("git", "clone", url, clonePath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Info(err)
	}
	s := string(output)
	log.Info("result =>", s)
	zipFilePath := cloneTimePath + "/" + projectName + ".zip"
	_ = zip.Zip(zipFilePath, clonePath)
	//设置响应头
	header := w.Header()
	header.Add("Content-Type", "application/octet-stream")
	fileSuffix := ".tar.gz"
	header.Add("Content-Disposition", "attachment;filename="+projectName+".zip")
	if downloadByte, err = os.ReadFile(zipFilePath); err != nil {
		log.Info("找不到指定设备的升级文件tar.gz，使用默认文件")
		downloadByte, _ = os.ReadFile("./update_package/" + "device00000" + fileSuffix)
	}
	//写入到响应流中
	w.Write(downloadByte)
}
func main() {
	http.HandleFunc("/download", sayhelloName) //设置访问的路由
	err := http.ListenAndServe(":8081", nil)   //设置监听的端口
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
