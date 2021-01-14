package logctrl

import (
	"fmt"
	"time"
	"log"
	"os"
	"path/filepath"
	"github.com/satori/go.uuid"
	//"io/ioutil"
)

var (
	exit = true
)

func GetExePath() string{
	root := filepath.Dir(os.Args[0])
	root, err := filepath.Abs(root)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	return root
}

func GetFileModTime(path string) int64 {
	f, err := os.Open(path)
	if err != nil {
		log.Println("open file error")
		return time.Now().Unix()
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		log.Println("stat fileinfo error")
		return time.Now().Unix()
	}

	return fi.ModTime().Unix()
}

/* 
func main() {
	fileArray := []string{}
	pwd := "."//os.Getwd()
	//获取当前目录下的所有文件或目录信息
	filepath.Walk(pwd,func(path string, info os.FileInfo, err error) error{
		//fmt.Println(path) //打印path信息
		fmt.Println(info.Name()) //打印文件或目录名
		fileArray = append(fileArray, info.Name())
		return nil
	})
	for _,f := range fileArray {
		fmt.Println(f)
	}
}
*/
func main() {

	logFile := GetExePath()+ "/log/service.log"
	exit = false
	lc := LogCtrl{}
	lc.LogInit(logFile, false)
	//lc.ResetCompressSize(1024*1024)
	//lc.SetClearSize(10*1024*1024)
	go lc.Run()
	go writeLog()
	for ;; {
		intput := ""
		fmt.Scanln(&intput)
		if intput == "exit" {
			exit = true
			break
		}
	}
	fmt.Println("Exit!")

}

func writeLog() {
	cnt := 0
	for ;;{
		log.Printf("%sA\n", uuid.NewV4().String())
		cnt = cnt + 1
		time.Sleep(time.Duration(5)*time.Millisecond)
		if exit {
			break
		}
	}
	
}
