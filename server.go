package main

import (
	"fmt"
        "flag"
	"io"
        "log"
        "net/http"
	"os"
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
		s := db.Get(key)
		if s != "" {
			http.Redirect(w,req, s, 301)
		} else {
			http.Redirect(w,req, "https://google.com", 301)
		}
	}

        put := func(w http.ResponseWriter, req *http.Request) {
		switch {
		case req.Method == "GET": {
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
			value := req.Form.Get("value")

			if key == "create" {
				fmt.Fprintf(w, "I saw what you did :P")
			} else {
				db.Put(key, value)
				http.Redirect(w, req, value, 301)
			}
		}
		}
	}

	http.HandleFunc("/", get)
	http.HandleFunc("/create", put)

        err := http.ListenAndServe(":8080", nil)
        if err != nil {
                log.Fatal("ListenAndServe: ", err)
        }
}