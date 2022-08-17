package view

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"wb/tasks/lzero/db"
	"wb/tasks/lzero/models"
)

type Server struct {
	http.Server
	db           *db.Db
	viewTemplate *template.Template
}

var (
	//go:embed view.gohtml
	viewGohtml string
)

func Create(address string, db *db.Db) (*Server, error) {
	if db == nil {
		return nil, errors.New("Nil pointer to database")
	}

	viewTemplate, err := template.New("index").Parse(viewGohtml)
	if err != nil {
		return nil, err
	}

	mux := http.NewServeMux()

	server := Server{
		Server: http.Server{
			Addr:    address,
			Handler: mux,
		},
		db:           db,
		viewTemplate: viewTemplate,
	}

	mux.HandleFunc("/", server.handleRequest)

	go func() {
		err := server.ListenAndServe()
		if err != nil {
			fmt.Println(err)
		}
	}()

	return &server, nil
}

func (server *Server) Destroy() {
	server.Shutdown(context.Background())
}

func (server *Server) handleRequest(w http.ResponseWriter, req *http.Request) {
	orderUid := req.URL.Query().Get("order_uid")

	orderJson, err := server.db.Get(orderUid)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404"))
		return
	}

	var order models.Order
	err = json.Unmarshal([]byte(orderJson), &order)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500"))
		return
	}

	err = server.viewTemplate.Execute(w, order)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500"))
		return
	}
	fmt.Printf("Responded to request for order_uid=%s\n", orderUid)
}
