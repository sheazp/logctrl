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
