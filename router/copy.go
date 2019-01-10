package router

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/blueberryserver/bluecore/blueprocess"

	"github.com/julienschmidt/httprouter"
)

// Config info
type Config struct {
	Host     string   `json:"host"`
	Port     string   `json:"port"`
	User     string   `json:"user"`
	Pw       string   `json:"pw"`
	Database []string `json:"database"`
	Dump     []string `json:"dump"`
	Option   []string `json:"option"`
	Path     string   `json:"path"`
}

func dbExport(config Config) error {

	for index, database := range config.Database {
		fmt.Printf("dump database:%s -> dump:%s\r\n", database, config.Dump[index])

		dumpfile := fmt.Sprintf("%s//%s", config.Path, config.Dump[index])

		option := []string{
			"--skip-opt", "--net_buffer_length=409600", "--create-options", "--disable-keys",
			"--routines", "--triggers", "--lock-tables", "--quick", "--set-charset", "--extended-insert", "--single-transaction", "--add-drop-table",
			"--no-create-db"}

		for _, op := range config.Option {
			option = append(option, op)
		}

		{
			option = append(option, "-h")
			option = append(option, config.Host)
			option = append(option, "--port")
			option = append(option, config.Port)
			option = append(option, "-u")
			option = append(option, config.User)
			option = append(option, config.Pw)
			option = append(option, database)
		}
		fmt.Println(option)

		err := blueprocess.Execute("mysqldump.exe", dumpfile, option...)

		if err != nil {
			return err
		}
	}
	return nil
}

func dbCreate(config Config) error {
	for _, database := range config.Database {

		fmt.Printf("create %s\r\n", database)

		err := blueprocess.Execute("mysqladmin.exe", "", "-h", config.Host, "--port", config.Port, "-u", config.User, config.Pw, "create", database)
		if err != nil {
			return err
		}
	}
	return nil
}

func dbImport(config Config) error {

	dbNames := make([]string, 0)
	for index, database := range config.Database {
		dumpfile := fmt.Sprintf("%s//%s", config.Path, config.Dump[index])
		fmt.Printf("import database:%s <- dump:%s\r\n", database, config.Dump[index])

		f, err := os.Open(dumpfile)
		defer f.Close()
		if err != nil {
			fmt.Println("file open errer")
			return err
		}

		dump, err := ioutil.ReadAll(f)
		if err != nil {
			fmt.Println("file open errer")
			return err
		}
		err = blueprocess.ExecutePipeIn("mysql.exe", string(dump), "-h", config.Host, "--port", config.Port, "-u", config.User, config.Pw, database)

		if err != nil {
			return err
		}
		fmt.Println("Complete", dumpfile)

		dbNames = append(dbNames, database)
	}

	//shardIdxs := []int{1, 101, 102, 301}
	//settingShardInfo(GDB, shardIdxs, dbNames)
	return nil
}

// Export ...
/*
{
	"host":"0.0.0.0",
	"port":"3306",
	"user":"user",
	"pw":"-ppasswd",
	"database": [
		"setting_db",
		"user_shard1_db",
		"user_shard2_db",
		"log_db"
	],
	"dump": [
		"setting_db_ub25.sql",
		"user_shard1_db_ub25.sql",
		"user_shard2_db_ub25.sql",
		"log_db_ub25.sql"
	],
	"option": [],
	"path":"copy/ub25/files"
}
*/
func Export(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	decoder := json.NewDecoder(r.Body)

	var config Config
	err := decoder.Decode(&config)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	fmt.Println(config)
	return
	dbExport(config)
}

// Create ...
/*
{
	"host":"211.218.232.151",
	"port":"3306",
	"user":"doz2",
	"pw":"-ptest1324",
	"database": [
		"setting_db_ub25",
		"user_shard1_db_ub25",
		"user_shard2_db_ub25",
		"log_db_ub25"
	]
}
*/
func Create(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	decoder := json.NewDecoder(r.Body)

	var config Config
	err := decoder.Decode(&config)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	fmt.Println(config)
	return
	dbCreate(config)
}

// Import ...
/*
{
	"host":"211.218.232.151",
	"port":"3306",
	"user":"doz2",
	"pw":"-ptest1324",
	"database": [
		"setting_db_ub25",
		"user_shard1_db_ub25",
		"user_shard2_db_ub25",
		"log_db_ub25"
	],
	"dump": [
		"setting_db_ub25.sql",
		"user_shard1_db_ub25.sql",
		"user_shard2_db_ub25.sql",
		"log_db_ub25.sql"
	],
	"path":"copy/ub25/files"
}
*/
func Import(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	decoder := json.NewDecoder(r.Body)

	var config Config
	err := decoder.Decode(&config)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	fmt.Println(config)
	return
	dbImport(config)
}
