package router

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/blueberryserver/bluecore"
	"github.com/blueberryserver/bluecore/bluecp"
	"github.com/blueberryserver/bluecore/bluemerge"
	"github.com/blueberryserver/bluecore/blueprocess"
	"github.com/julienschmidt/httprouter"
)

type database struct {
	Name   string   `json:"name"`
	Tables []string `json:"tables"`
}

type dumpConfig struct {
	Host  string `json:"host"`
	User  string `json:"user"`
	Pw    string `json:"pw"`
	Line  string `json:"line"`
	Path  string `json:"path"`
	DbVer string `json:"dbver"`
}

// Context ...
type Context struct {
	Database database   `json:"database"`
	Config   dumpConfig `json:"config"`
}

func dump(tables []string, config dumpConfig, database string) error {
	messageChan := make(chan string)
	var wg sync.WaitGroup
	wg.Add(len(tables))

	var filenames = make([]string, 100)
	for i, table := range tables {

		var filename = config.Path + "/" + table + ".sql"
		var tempfilename = config.Path + "/" + table + "_dump.sql"

		go func(tablename string, name string, tempname string) {
			defer wg.Done()

			// mysqmdump.exe 실행 테입블 덤프
			err := blueprocess.Execute("./bin/mysqldump.exe", tempname, "--skip-opt", "--net_buffer_length=409600", "--create-options", "--disable-keys",
				"--lock-tables", "--quick", "--set-charset", "--extended-insert", "--single-transaction", "--add-drop-table", "--no-create-db", "-h",
				config.Host, "-u", config.User, config.Pw, database, tablename)
			if err != nil {
				log.Println(err)
				messageChan <- name + " fail"
				return
			}

			if config.Line != "1" {
				// ),( -> )\n( 변경 작업 줄바꿈 처리
				err = blueprocess.Execute("./bin/sed.exe", name, "s/),(/),\\\\r\\\\n(/g", tempname)
				if err != nil {
					log.Println(err)
					messageChan <- name + " fail"
					return
				}
			} else {
				err = blueprocess.Execute("./bin/sed.exe", name, "", tempname)
				if err != nil {
					log.Println(err)
					messageChan <- name + " fail"
					return
				}
			}

			bluecp.RM(tempname)

			messageChan <- tablename + " success"
		}(table, filename, tempfilename)

		filenames[i] = filename
	}

	go func() {
		for msg := range messageChan {
			log.Printf("%s\r\n", msg)
		}
	}()

	wg.Wait()

	Now := time.Now()
	// layout (2006-01-02 15:04:05)
	temp := Now.Format("20060102-150405")
	//fmt.Println("Now: ", temp)
	var resultfilename = config.Path + "/" + database + "_" + temp + ".sql"

	//merge.MERGE(resultfilename, config.Tables...)
	bluemerge.MERGE(resultfilename, filenames...)
	bluecore.ZipFiles(config.Path+"\\"+database+"_"+temp+".zip", filenames)
	for _, file := range filenames {
		bluecp.RM(file)
	}

	//downloadFile(database+".zip", "http://localhost:8080/"+database+".zip")
	return nil
}

// Dump ...
/*
// sample json
{
	"config": {
		"host":"0.0.0.0",
		"user":"user",
		"pw":"passwd",
		"line":"2",
		"path":"dump/ub25/files",
		"dbver":"ub25"
	},

	"database": {
		"name":"doz3_global_ub25",
		"tables":[
			"setting_table",
		]
	}
}
*/
func Dump(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	decoder := json.NewDecoder(r.Body)
	var context = &Context{}
	err := decoder.Decode(context)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), 400)
		return
	}

	fmt.Println(context)
	return
	err = dump(context.Database.Tables, context.Config, context.Database.Name)
	if err == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)

		// Result info
		type Result struct {
			ResultCode string `json:"resultcode"`
		}
		result := Result{"0"}
		json.NewEncoder(w).Encode(result)

	} else {
		log.Println(err)
		http.Error(w, err.Error(), 400)
		return
	}
}

// FileContext ...
type FileContext struct {
	Path     string   `json:"path"`
	Files    []string `json:"files"`
	SaveData string   `json:"data"`
}

// Delete ...
// sample json
/*
{
	"path":"dump/ub25/files",
	"files":[
		"setting_table.sql"
	],
	"data":""
}
*/
func Delete(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	decoder := json.NewDecoder(r.Body)

	var context = &FileContext{}
	err := decoder.Decode(context)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), 400)
		return
	}

	fmt.Println(context)
	return
	// 서버 파일
	files, _ := filepath.Glob(context.Path)
	for _, p := range files {
		if strings.Contains(p, "selectedTable.txt") {
			continue
		}

		// 파일 존재 여부 확인
		exist := func(list []string, item string) bool {
			for _, v := range list {
				if v == item {
					return true
				}
			}
			return false
		}(context.Files, p)

		if true == exist {
			err := os.Remove(p)
			if err != nil {
				http.Error(w, err.Error(), 400)
				return
			}
		}
	}

	// Result info
	type Result struct {
		ResultCode string `json:"resultcode"`
	}
	result := Result{"0"}
	json.NewEncoder(w).Encode(result)
}

// Save ...
/*
{
	"path":"dump/ub25/etc",
	"files":[
		"saveData.txt"
	],
	"data":""
}
*/
func Save(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	decoder := json.NewDecoder(r.Body)
	var context = &FileContext{}
	err := decoder.Decode(context)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), 400)
		return
	}

	fmt.Println(context)
	return
	// remove file
	if len(context.Files) == 0 {
		http.Error(w, err.Error(), 400)
		return
	}

	path := context.Path + "/" + context.Files[0]

	os.Remove(path)
	//save file
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println(err)
	}
	defer f.Close()
	f.WriteString(context.SaveData + "\r\n")

	// Result info
	type Result struct {
		ResultCode string `json:"resultcode"`
	}
	result := Result{"0"}
	json.NewEncoder(w).Encode(result)
}
