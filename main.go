package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	dbrouter "github.com/blueberryserver/bluedbapi/router"
	"github.com/rs/cors"

	"github.com/blueberryserver/bluecore"
	"github.com/julienschmidt/httprouter"
)

// Config ...
type Config struct {
	Host  string `json:"host"`
	User  string `json:"user"`
	Pw    string `json:"pw"`
	Line  string `json:"line"`
	Path  string `json:"path"`
	DbVer string `json:"dbver"`
}

var gconfig = &Config{}

// DBConfig ...
type DBConfig struct {
	Configs []Config `yaml:"conf"`
}

// Table ...
type Table struct {
	Global []string `yaml:"global"`
	User   []string `yaml:"user"`
	Log    []string `yaml:"log"`
}

// yaml
var gconfs = &DBConfig{}
var gtables = &Table{}
var gPort *string

func initLogFile() {
	// log path check

	// 로그 파일
	/*fileLog*/
	_, err := bluecore.InitLog("log/log_" + time.Now().Format("2006_01_02_15") + ".txt")
	//defer fileLog.Close()
	if err != nil {
		fmt.Printf("%s\n\n", err)
		return
	}
}

func loadConfig() {
	err := bluecore.ReadYAML(gconfs, "conf/conf.yaml")
	if err != nil {
		log.Println(err)
		return
	}

	err = bluecore.ReadYAML(gtables, "conf/tables.yaml")
	if err != nil {
		log.Println(err)
		return
	}
}

func loadFlag() {
	gPort = flag.String("p", "8080", "p=8080")
	flag.Parse() // 명령줄 옵션의 내용을 각 자료형별
	log.Printf("Start db api (%s) \r\n", *gPort)
}

func routing() {
	// start routing
	router := httprouter.New()

	router.POST("/dump", dbrouter.Dump)
	router.POST("/delete", dbrouter.Delete)
	router.POST("/save", dbrouter.Save)
	router.GET("/files/:ubver/:cmd", dbrouter.FileList)
	//router.ServeFiles()

	router.POST("/import", dbrouter.Import)
	router.POST("/create", dbrouter.Create)
	router.POST("/export", dbrouter.Export)
	router.ServeFiles("/files/:ubver/:cmd/*filepath", http.Dir("files/:ubver/:cmd"))

	handler := cors.AllowAll().Handler(router)
	log.Fatal(http.ListenAndServe(":"+*gPort, handler))
}

func main() {

	initLogFile()

	loadConfig()

	loadFlag()

	routing()
}
