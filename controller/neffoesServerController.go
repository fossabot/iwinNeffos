package controller

import (
	"log"
	"net/http"
	"strconv"

	"github.com/majidbigdeli/neffosAmi/domin/data"
	"github.com/majidbigdeli/neffosAmi/domin/variable"

	"github.com/kataras/neffos"
	"github.com/kataras/neffos/gobwas"
)

var (
	//Server ...
	Server      *neffos.Server
	connections = make(map[string]*neffos.Conn)
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

		connections[c.ID()] = c

		if c.WasReconnected() {
			log.Printf("[%s] connection is a result of a client-side re-connection, with tries: %d", c.ID(), c.ReconnectTries)
		}
		log.Printf("[%s] connected to the server.", c)
		return nil
	}

	server.OnDisconnect = func(c *neffos.Conn) {
		delete(connections, c.ID())
		log.Printf("[%s] disconnected from the server.", c)
	}

	Server = server
}

var events = neffos.Namespaces{
	variable.Agent: neffos.Events{
		// ایونت متصل شدن به یک فضای نام
		neffos.OnNamespaceConnected: func(c *neffos.NSConn, msg neffos.Message) error {
			log.Printf("[%s] connected to namespace [%s].", c, msg.Namespace)
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
	},
}
