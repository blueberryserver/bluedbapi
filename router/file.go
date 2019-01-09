package router

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/julienschmidt/httprouter"
)

type fileInfo struct {
	Name string    `json:"name"`
	Size int64     `json:"size"`
	Time time.Time `json:"time"`
}

// FileList ...
func FileList(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	//v := r.URL.Query()

	ubVer := ps.ByName("ubver")
	cmd := ps.ByName("cmd")
	// url path
	path := r.URL.Path
	log.Printf("path=%s ver=%s cmd=%s", path, ubVer, cmd)

	var files []fileInfo
	filepath.Walk("."+path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			log.Println(err)
		}
		if false == info.IsDir() {
			files = append(files, fileInfo{info.Name(), info.Size(), info.ModTime()})
			//_, name := filepath.Split(p)
			//files = append(files, name)
		}
		return nil
	})

	log.Printf("%v", files)

	// Result info
	type Result struct {
		ResultCode string     `json:"resultcode"`
		FileList   []fileInfo `json:"files"`
	}
	result := Result{"0", files}
	json.NewEncoder(w).Encode(result)
	return
}
