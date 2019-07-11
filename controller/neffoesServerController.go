package controller

import (
	"log"
	"net/http"

	"github.com/kataras/neffos"
	"github.com/kataras/neffos/gobwas"
	"github.com/majidbigdeli/neffosAmi/event"
)

//Server ...
var Server *neffos.Server

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

		// if returns non-nil error then it refuses the client to connect to the server.
		return nil
	}

	server.OnDisconnect = func(c *neffos.Conn) {
		log.Printf("[%s] disconnected from the server.", c)
	}

	Server = server
}
