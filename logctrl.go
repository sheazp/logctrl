package logctrl 

import (
	"os"
	"time"
	"io"
	"strings"
	"github.com/mholt/archiver"
	"log"
	"fmt"
	"path/filepath"
)

type LogCtrl struct {
	FileName   		string    // 原始文件名
	CompressMethod  string    // 压缩方式
	Directory       string    // 文件路径
	TrigerSize      int64     // 触发压缩的大小
	LogName         string    // 日志服务名
	CompresCnt      int32
	StdOut          bool      // 是否终端打印日志
	AllZipMaxSize   int64     // 压缩文件最大总大小
	ZipMaxCount     int64     // 压缩文件最大个数
}
func getFileSize(FileName string) int64 {
	fi,err:=os.Stat(FileName)
    if err !=nil {
		return -1
	}
	//fmt.Println("file size is ",fi.Size(),err)
	return  fi.Size()
}

func fileRename(src_file, dst_file string) bool {
    err := os.Rename(src_file, dst_file)     //重命名
    if err != nil {
        //如果重命名文件失败,则输出错误 file rename Error!
        fmt.Println("[Logctrl] file rename Error:", err)
		return false
    }
	return true
}

func fileZip(src_file, zip_file string) bool {
	// 压缩文件
	err := archiver.Archive( []string{src_file},zip_file)
	if err != nil {
		fmt.Println("[Logctrl] File zip fail,err:", err)
		return false
	}
	err = os.Remove(src_file)               //删除文件
	if err != nil {
		fmt.Println("[Logctrl] File remove fail: ", err)
		return false
	}
	return true
}
func (this *LogCtrl) resetLogWriter(){
	file, _ := os.OpenFile(this.FileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	writers := []io.Writer{
		file}
	if this.StdOut {
		writers = append(writers, os.Stdout)
	}
	fileAndStdoutWriter := io.MultiWriter(writers...)
	log.SetOutput(fileAndStdoutWriter)
}
func  (this *LogCtrl)Init(FileName string) {
	 this.LogInit(FileName ,true)
}
func (this *LogCtrl)LogInit(FileName string,StdOut bool){
	if len(FileName) == 0 {
		FileName = "default.log"
		this.Init(FileName)
	}
	this.StdOut = StdOut
	this.FileName = FileName
	this.CompressMethod = "zip"   // 默认zip格式
	this.TrigerSize = 20*1024*1024 // 默认10MB进行压缩
	this.AllZipMaxSize = 100*1024*1024 // 默认压缩大小超过100MB，删除最前的日志
	this.ZipMaxCount = 30  // 默认压缩包个数最大20
	n := strings.LastIndexByte(FileName, '/')
	if n < 0 {
		this.Directory = "."   // 当前文件夹
		n = 0
	}else {
		this.Directory = FileName[0:n]
	}
	s := strings.LastIndexByte(FileName, '.')
	if s < 0 {
		this.LogName = FileName[n: len(FileName)]
	}else {
		if s <= n {
			this.LogName = FileName
		}else{
			this.LogName = FileName[n+1:s]
		}
	}
	this.resetLogWriter()
	fmt.Println("LogFile: ", this.FileName)
	fmt.Println("LogDir: ", this.Directory)
	fmt.Println("LogMeth: ", this.CompressMethod)
	fmt.Println("LogName: ", this.LogName)

}
func (this *LogCtrl)Run(){
	if len(this.LogName) == 0 {
		this.Init("default.log")
	}
	this.resetLogWriter()
	tickToCheckClear := 0
	this.doclear()
	for ;; {
		filsize := getFileSize(this.FileName)
		if filsize > this.TrigerSize {
			//fmt.Println("Start zip..")
			header   := this.LogName + ".log@" + time.Now().Format("20060102_150405")
			newlfile := this.Directory + "/" +header + ".log"
			zipFile  := this.Directory + "/" +header + "." + this.CompressMethod
			if fileRename(this.FileName, newlfile) {
				this.resetLogWriter()
				fileZip(newlfile, zipFile)
				this.CompresCnt += 1
			}
		}
		time.Sleep(time.Duration(1)*time.Second)
		tickToCheckClear++
		//定期检查是否需要清除zip文件
		if tickToCheckClear > 5 {
			this.doclear()
			tickToCheckClear = 0
		}
	}
}

func (this *LogCtrl)ResetCompressSize(s int64/*字节*/){
	if s < 100*1024 {
		return 
	}
	this.TrigerSize = s
	log.Println("[Logctrl] Logctrl TriggerSize = ", this.TrigerSize)
}
func (this *LogCtrl)SetClearSize(s int64 /*字节*/){
	if s < 100*1024 {
		return 
	}
	this.AllZipMaxSize = s
	log.Println("[Logctrl] Logctrl AllZipMaxSize = ", this.AllZipMaxSize)
}
func (this *LogCtrl)SetZipMaxCount(count int64 ){
	if count < 30 {
		return 
	}
	this.ZipMaxCount = count
	log.Println("[Logctrl] Logctrl ZipMaxCount = ", this.ZipMaxCount)
}

func getFileModTime(path string) int64 {
	fi, err := os.Stat(path)
	if err != nil {
		log.Println("stat fileinfo error")
		return time.Now().Unix()
	}

	return fi.ModTime().Unix()
}

type zipfile struct {
	FileName          string
	PathName          string
	LastModifyTime    int64
	FileSize          int64
}

func (this *LogCtrl) doclear(){
	fileArray := []string{}
	//zipFInfo := make([]zipfile, 0)
	pwd := this.Directory
	//获取当前目录下的所有文件或目录信息
	filepath.Walk(pwd,func(path string, info os.FileInfo, err error) error{
		if err == nil {
			//fmt.Println(path) //打印path信息
			//fmt.Println(info.Name()) //打印文件或目录名
			if fileArray != nil {
				fileArray = append(fileArray, info.Name())
			}
		}
		return nil
	})
	delTarget := ""
	zipCount := int64(0)
	targetModTime := time.Now().Unix()
	zipSize := int64(0)
	for _,f := range fileArray {
		if !strings.HasPrefix(f, this.LogName + ".log@") {
			//fmt.Printf("f %s, not contain %s\n", f, this.LogName)
			continue 
		}
		if !strings.HasSuffix(f, "." +this.CompressMethod ) {
			//fmt.Printf("f %s, not contain %s\n", f, "." + this.CompressMethod)
			continue
		}
		z := zipfile{}
		z.FileName = f
		z.PathName = this.Directory+ "/"+f
		z.LastModifyTime = getFileModTime(z.PathName)
		z.FileSize = getFileSize(z.PathName)
		zipSize = zipSize + z.FileSize
		if targetModTime > z.LastModifyTime {
			targetModTime = z.LastModifyTime
			delTarget = z.PathName
		}
		zipCount ++
		//zipFInfo = append(zipFInfo, z)
	}
	//fmt.Printf("TotalZie: %d , MaxSize: %d\n", zipSize, this.AllZipMaxSize)
	if zipSize > this.AllZipMaxSize  || zipCount > this.ZipMaxCount{
		if len(delTarget) > 0  {
			//每次只删除
			fmt.Println("[logctrl]clear log file:", delTarget)
			err := os.Remove(delTarget)               //删除文件
			if err != nil {
				log.Println("[Logctrl] File remove fail: ", err)
				return 
			}
		}
	}
	return 
}
