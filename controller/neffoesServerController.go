package controller

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/kataras/neffos"
	"github.com/kataras/neffos/gobwas"
	"github.com/majidbigdeli/neffosAmi/event"
)

var (
	Server           *neffos.Server
	connections      = make(map[string]*neffos.Conn)
	connectionIDs    []string
	connID           = make(map[string]string)
	addConnection    = make(chan *neffos.Conn)
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

		addConnection <- c

		if c.WasReconnected() {
			log.Printf("[%s] connection is a result of a client-side re-connection, with tries: %d", c.ID(), c.ReconnectTries)
		}

		log.Printf("[%s] connected to the server.", c)

		// if returns non-nil error then it refuses the client to connect to the server.
		return nil
	}

	server.OnDisconnect = func(c *neffos.Conn) {

		removeConnection <- c

		log.Printf("[%s] disconnected from the server.", c)
	}

	Server = server
}

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

func StartConnectionManager(ctx context.Context) {
	if ctx == nil {
		ctx = context.TODO()
	}

	go func() {
		for {
			select {
			case c := <-addConnection:
				connections[c.ID()] = c
				connectionIDs = append(connectionIDs, c.ID())
			case c := <-removeConnection:
				delete(connections, c.ID())
				delete(connID, c.ID())

				if len(connectionIDs) == 1 {
					connectionIDs = connectionIDs[0:0]
				} else {
					for i, n := 0, len(connectionIDs); i < n; i++ {
						if connectionIDs[i] == c.ID() {
							connectionIDs = append(connectionIDs[0:i], connectionIDs[i+1:]...)
							break
						}
					}
				}

			case <-ctx.Done():
				return
			}
		}
	}()
}

type Response struct {
	Status bool
}
