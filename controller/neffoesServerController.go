package controller

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/kataras/neffos"
	"github.com/kataras/neffos/gobwas"
	"github.com/majidbigdeli/neffosAmi/event"
)

var (
	//Server ...
	Server           *neffos.Server
	connID           = make(map[string]string)
	removeConnection = make(chan *neffos.Conn)
	mux              sync.Mutex
)

//NeffosServer For neffos
func NeffosServer() {
	server := neffos.New(gobwas.DefaultUpgrader, event.Events)

	server.IDGenerator = func(w http.ResponseWriter, r *http.Request) string {
		if extension := r.Header.Get("Extension"); extension != "" {
			return extension
		}
		return neffos.DefaultIDGenerator(w, r)
	}

	server.OnUpgradeError = func(err error) {
		log.Printf("ERROR: %v", err)
	}

	server.OnConnect = func(c *neffos.Conn) error {
		if c.WasReconnected() {
			log.Printf("[%s] connection is a result of a client-side re-connection, with tries: %d", c.ID(), c.ReconnectTries)
		}
		log.Printf("[%s] connected to the server.", c)
		return nil
	}

	server.OnDisconnect = func(c *neffos.Conn) {
		removeConnection <- c
		log.Printf("[%s] disconnected from the server.", c)
	}

	Server = server
}

//Check ....
func Check(w http.ResponseWriter, r *http.Request) {
	connectionID := r.URL.Query().Get("connectionId")

	mux.Lock()
	_, ok := connID[connectionID]

	if ok {

		response := Response{false}

		toReturn, err := json.Marshal(response)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(toReturn)

	} else {
		connID[connectionID] = connectionID
		response := Response{true}
		toReturn, err := json.Marshal(response)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(toReturn)
	}
	mux.Unlock()

}

//StartConnectionManager  ....
func StartConnectionManager() {
	go func() {
		for {
			select {
			case c := <-removeConnection:
				delete(connID, c.ID())
			}
		}
	}()
}

//Response ...
type Response struct {
	Status bool
}
