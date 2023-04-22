package jsonapi

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log"
	"mailing/mdb"
	"net/http"
)

func setJsonHeader(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charser=utf-8")
}

func fromJson[T any](body io.Reader, target T) {
	buf := new(bytes.Buffer)
	buf.ReadFrom(body)
	json.Unmarshal(buf.Bytes(), &target)
}

func returnJson[T any](w http.ResponseWriter, withData func() (T, error)) {
	setJsonHeader(w)

	data, serverErr := withData()
	if serverErr != nil {
		w.WriteHeader(500)
		serverErrJson, err := json.Marshal(&serverErr)
		if err != nil {
			log.Print(err)
			return
		}
		w.Write(serverErrJson)
	}
	dataJson, err := json.Marshal(&data)
	if err != nil {
		log.Println(err)
		w.WriteHeader(500)
		return
	}

	w.Write(dataJson)
}

func returnErr(w http.ResponseWriter, err error, code int) {
	returnJson(w, func() (interface{}, error) {
		errorMessage := struct {
			Err string
		}{
			Err: err.Error(),
		}
		w.WriteHeader(code)
		return errorMessage, nil
	})
}

func CreateEmail(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != "POST" {
			return
		}
		entry := mdb.EmailEntry{}
		fromJson(req.Body, &entry)
		if err := mdb.CreateEmail(db, entry.Email); err != nil {
			returnErr(w, err, 400)
			return
		}
		returnJson(w, func() (interface{}, error) {
			log.Printf("Jsom createEmail: %v\n", entry.Email)
			return mdb.GetEmail(db, entry.Email)
		})
	})
}

func GetEmail(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != "GET" {
			return
		}
		entry := mdb.EmailEntry{}
		fromJson(req.Body, &entry)
		returnJson(w, func() (interface{}, error) {
			log.Println("Jsom GetEmail: %v\n", entry.Email)
			return mdb.GetEmail(db, entry.Email)
		})
	})
}

func UpdateEmail(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != "PUT" {
			return
		}
		entry := mdb.EmailEntry{}
		fromJson(req.Body, &entry)
		if err := mdb.UpdateEmail(db, entry); err != nil {
			returnErr(w, err, 400)
			return
		}
		returnJson(w, func() (interface{}, error) {
			log.Printf("Jsom.UpdateEmail(): %v\n", entry.Email)
			return mdb.GetEmail(db, entry.Email)
		})
	})
}

func DeleteEmail(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != "POST" {
			return
		}
		entry := mdb.EmailEntry{}
		fromJson(req.Body, &entry)
		if err := mdb.Delete(db, entry.Email); err != nil {
			returnErr(w, err, 400)
			return
		}
		returnJson(w, func() (interface{}, error) {
			log.Printf("Jsom.Delete: %v\n", entry.Email)
			return mdb.GetEmail(db, entry.Email)
		})
	})
}

func GetBatchEmail(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != "GET" {
			return
		}
		queryOption := mdb.GetEmailBatchQueryParams{}
		fromJson(req.Body, &queryOption)
		if queryOption.Count <= 0 || queryOption.Page <= 0 {
			returnErr(w, errors.New("page and count were invalid"), 400)
			return
		}
		returnJson(w, func() (interface{}, error) {
			log.Printf("JSON GetBatchEmail: %v\n", queryOption)
			return mdb.GetEmailBatch(db, queryOption)
		})
	})
}

func Serve(db *sql.DB, bind string) {
	http.Handle("/email/create", CreateEmail(db))
	http.Handle("/email/get", GetEmail(db))
	http.Handle("/email/get_batch", GetBatchEmail(db))
	http.Handle("/email/delete", DeleteEmail(db))
	http.Handle("/email/update", UpdateEmail(db))
	log.Printf("JSON API serve listening on %v\n", bind)
	err := http.ListenAndServe(bind, nil)
	if err != nil {
		log.Fatalf("JSON SERVER ERROR: %v", err)
	}
}
