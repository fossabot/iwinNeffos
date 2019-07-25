package controller

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/majidbigdeli/neffosAmi/domin/data"
	"github.com/majidbigdeli/neffosAmi/domin/variable"

	"github.com/kataras/neffos"
	"github.com/kataras/neffos/gobwas"
)

type command struct {
	connectionID string
	replyChan    chan bool
}

var (
	//Server ...
	Server           *neffos.Server
	connID           = make(chan command)
	connections      = make(map[string]*neffos.Conn)
	addConnection    = make(chan *neffos.Conn)
	removeConnection = make(chan *neffos.Conn)
	nsConn           *neffos.NSConn
)

//NeffosServer For neffos
func NeffosServer() {
	server := neffos.New(gobwas.DefaultUpgrader, events)

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
	c := r.URL.Query().Get("connectionId")

	con, ok := connections[c]
	if ok {
		if con.IsClosed() {
			delete(connections, con.ID())
			response := Response{true}
			toReturn := neffos.Marshal(response)
			w.Header().Set("Content-Type", "application/json")
			w.Write(toReturn)
		} else {
			response := Response{false}
			toReturn := neffos.Marshal(response)
			w.Header().Set("Content-Type", "application/json")
			w.Write(toReturn)
		}
	} else {
		response := Response{true}
		toReturn := neffos.Marshal(response)
		w.Header().Set("Content-Type", "application/json")
		w.Write(toReturn)
	}

}

//Disconnect ....
func Disconnect(w http.ResponseWriter, r *http.Request) {
	c := r.URL.Query().Get("number")

	con, ok := connections[c]

	if ok {
		if con.IsClosed() {
			delete(connections, con.ID())
		} else {
			con.Close()
			delete(connections, con.ID())
		}
	}

	w.WriteHeader(http.StatusOK)
}

//StartConnectionManager  ....
func StartConnectionManager() {
	go func() {
		tick := time.Tick(5 * time.Second)
		for {
			select {
			case c := <-addConnection:
				con, ok := connections[c.ID()]
				if !ok {
					c.Connect(nil, variable.Agent)
				} else {
					if con.IsClosed() {
						c.Connect(nil, variable.Agent)
					}
				}
				connections[c.ID()] = c
			case c := <-removeConnection:
				delete(connections, c.ID())
			case <-tick:
				for k, v := range connections {
					if v.IsClosed() {
						delete(connections, k)
					}
				}

			}
		}
	}()
}

//Response ...
type Response struct {
	Status bool
}

var events = neffos.Namespaces{
	variable.Agent: neffos.Events{
		// ایونت متصل شدن به یک فضای نام
		neffos.OnNamespaceConnected: func(c *neffos.NSConn, msg neffos.Message) error {
			log.Printf("[%s] connected to namespace [%s].", c, msg.Namespace)
			nsConn = c

			id, _ := strconv.Atoi(c.Conn.ID())
			notification, err := data.GetNotifByUserID(id)
			if err != nil {
				return err
			}
			c.Conn.Server().Broadcast(nil, neffos.Message{
				Namespace: msg.Namespace,
				Event:     "resiveErja",
				To:        c.Conn.ID(),
				Body:      neffos.Marshal(notification),
			})
			return nil
		},

		// دیسکانکت شدن از فضای نام
		neffos.OnNamespaceDisconnect: func(c *neffos.NSConn, msg neffos.Message) error {
			log.Printf("[%s] disconnected from namespace [%s].", c, msg.Namespace)
			return nil
		},

		// ایونت آپدیت کردن وضعیت اکستنشن
		variable.OnUpdateStatusNotification: func(c *neffos.NSConn, msg neffos.Message) error {
			id, _ := strconv.Atoi(string(msg.Body))
			data.UpdateNotification(id)
			return nil
		},
		"erja": func(c *neffos.NSConn, msg neffos.Message) error {

			useriDs := string(msg.Body)
			id, _ := strconv.Atoi(useriDs)

			notification, err := data.GetNotifByUserID(id)

			if err != nil {
				return err
			}

			c.Conn.Server().Broadcast(nil, neffos.Message{
				Namespace: msg.Namespace,
				Event:     "resiveErja",
				To:        useriDs,
				Body:      neffos.Marshal(notification),
			})
			return nil
		},
	},
}
