package main

import (
	"fmt"
	"net/http"
	"os"
	"tdirectory/common"
	hndlr "tdirectory/handler"
	"time"

	"github.com/gorilla/mux"
)

type tdr common.Server

func main() {
	tele := newDirectory()
	tele.initialiseDirectory()
}

func newDirectory() *tdr {
	t := new(tdr)
	t.Name = "Sample Telephone Directory"
	t.Address = os.Getenv("IP_N_PORT")
	if t.Address == "" {
		t.Address = ":8080"
	}
	t.Router = mux.NewRouter().StrictSlash(true)
	t.updateRoutes()
	return t
}

func (t *tdr) initialiseDirectory() {
	for {
		e := http.ListenAndServe(t.Address, t.Router)
		if e != nil {
			fmt.Println("Error inititalising telephone directory", e.Error())
			time.Sleep(10 * time.Second)
			continue
		}
		break
	}
}

func (t *tdr) updateRoutes() {
	t.Router.StrictSlash(true)
	t.Router.HandleFunc("/", serverStarted)
	for _, r := range hndlr.DirectoryRoutes {
		t.Router.Methods(r.Method).Path(r.Path).Name(r.Name).Handler(r.Handler)
	}
}

func serverStarted(w http.ResponseWriter, r *http.Request) {
	_, _ = fmt.Fprintf(w, "Hello there mate\n")
}
