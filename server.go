package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
        "flag"
        "log"
        "net/http"
	"strings"
)

import  "github.com/gwenn/gosqlite"

type Db  struct {
	*sqlite.Conn
}

func (db *Db) Put(key string, value string) {
	ins, err := db.Prepare("INSERT INTO url (key, value) VALUES (?,?)")
	if err != nil {
		log.Println("Error Put: ", err)
	}
	defer ins.Finalize()
	_, err = ins.Insert(key,value)
	if err != nil {
		log.Println("Error Insert: ", err)
	}
}

func (db *Db) Get(key string) string {
	s, err := db.Prepare("SELECT value from url WHERE key = ?", key)
	if err != nil {
		log.Println("Error: Get failed")
		return ""
	}
	defer s.Finalize()

	var value string
	retrieveKey := func(s *sqlite.Stmt) (err error) {
		if err = s.Scan(&value); err != nil {
			return err
		}
		return nil
	}
	_ = s.Select(retrieveKey)
	return value
}

func (db *Db) Delete(key string) {
	del, err := db.Prepare("DELETE FROM url WHERE key = ?")
	if err != nil {
		log.Println("Error: delete failed")
	}
	defer del.Finalize()

	_, err = del.ExecDml(key)
	if err != nil {
		log.Println("Error Insert: ", err)
	}

}


func (db *Db) GetAll() map[string]string {
	s, _ := db.Prepare("SELECT key,value from url")
	defer s.Finalize()

	all := make(map[string]string)
	getall := func(s *sqlite.Stmt) (err error) {
		n := s.ColumnCount()
		for i := 0 ; i < n ; i++ {
			var value string
			_, err := s.ScanByIndex(i, &value)
			if err != nil {
				fmt.Println(err)
				return err
			}
			key := value
			all[key] = ""
			_, err = s.ScanByIndex(i+1, &value)
			if err != nil {
				fmt.Println(err)
				return err
			}
			all[key] = value
			return nil
		}
		return nil
	}
	_ = s.Select(getall)
	return all
}
func InitDb(filename string) *Db {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		db, err := sqlite.Open(filename)
		if err != nil {
			log.Fatal("InitDb: ", err)
		}
		log.Println("Database doesn't exists, creating")
		err = db.Exec("CREATE TABLE url(key STRING PRIMARY KEY NOT NULL, value TEXT NOT NULL UNIQUE);")
		if err != nil {
			log.Fatal("Initdb -- create table url ", err)
		}
		return &Db{db}
	}
	db, err := sqlite.Open(filename)
	if err != nil {
		log.Fatal("InitDb: ", err)
	}
	log.Println("Found/Load previous initilised db")
	return &Db{db}
}

func main() {
        var filename = flag.String("filename", "goshrt.sqlite", "Sqlite database filename")
        flag.Parse()

	var db = InitDb(*filename)
	defer db.Close()

        get := func(w http.ResponseWriter, req *http.Request) {
		key := req.URL.Path
		key = strings.TrimLeft(key, "/")
		req.Body.Close()
		s := db.Get(key)
		if s != "" {
			http.Redirect(w,req, s, 301)
		} else {
			http.Redirect(w,req, "http://go/help", 301)
		}
	}

        put := func(w http.ResponseWriter, req *http.Request) {
		switch {
		case req.Method == "GET": {
			req.Body.Close()
			file, err := os.Open("create.html")
			if err != nil {
				log.Println("Error: ", err)
			}
			io.Copy(w,file)
		}
		case req.Method == "POST": {
			err := req.ParseForm()
			if err != nil {
				log.Println("CREATE POST: ", err)
			}

			key := req.Form.Get("key")
			if key == "" {
				newkey, _ := exec.Command("uuidgen").Output()
				key = strings.Replace(string(newkey), "\n", "",-1)
			}
			value := req.Form.Get("value")
			req.Body.Close()
			if key == "create" || key == "listall" || key == "delete" {
				fmt.Fprintf(w, "I saw what you did, abooooort! :P")
			} else if strings.Contains(value, "http://go/") {
				fmt.Fprintf(w, "I saw what you did, abooooort! :P")
			} else {
				db.Put(key, value)
				http.Redirect(w, req, value, 301)
			}
		}
		}
	}

	remove := func(w http.ResponseWriter, req *http.Request) {
		del := req.URL.Query().Get("key")
		db.Delete(del)
		http.Redirect(w,req, "https://google.com", 301)
	}

	listall := func(w http.ResponseWriter, req *http.Request) {
		allKeys := db.GetAll()
		jAll, _ := json.Marshal(allKeys)
		fmt.Fprintf(w, string(jAll))
	}

	help := func(w http.ResponseWriter, req *http.Request) {
		file, err := os.Open("help.html")
		if err != nil {
			log.Println("Error: ", err)
		}
		io.Copy(w,file)
	}

	http.Handle("/static/", http.FileServer(http.Dir("./")))

	http.HandleFunc("/", get)
	http.HandleFunc("/help", help)
	http.HandleFunc("/create", put)
	http.HandleFunc("/delete", remove)
	http.HandleFunc("/listall", listall)

        err := http.ListenAndServe(":8080", nil)
        if err != nil {
                log.Fatal("ListenAndServe: ", err)
        }
}
