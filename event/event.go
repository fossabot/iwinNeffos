package event

import (
	"log"
	"strconv"

	"github.com/kataras/neffos"
	"github.com/majidbigdeli/neffosAmi/domin/data"
	"github.com/majidbigdeli/neffosAmi/domin/variable"
)

//Events ...
var Events = neffos.Namespaces{
	variable.Agent: neffos.Events{
		// ایونت متصل شدن به یک فضای نام
		neffos.OnNamespaceConnected: func(c *neffos.NSConn, msg neffos.Message) error {
			log.Printf("[%s] connected to namespace [%s].", c, msg.Namespace)
			to := c.Conn.ID()
			id, err := data.GetUserIDByExtention(to)

			if err != nil {
				return nil
			}
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

			id, _ := strconv.Atoi(string(msg.Body))

			extention, err := data.GetExtentinByUserID(id)
			if err != nil {
				return err
			}

			notification, err := data.GetNotifByUserID(id)

			if err != nil {
				return err
			}

			ext := strconv.Itoa(extention)

			c.Conn.Server().Broadcast(nil, neffos.Message{
				Namespace: msg.Namespace,
				Event:     "resiveErja",
				To:        ext,
				Body:      neffos.Marshal(notification),
			})
			return nil
		},
	},
}
