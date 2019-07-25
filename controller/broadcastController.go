package controller

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/kataras/neffos"
	"github.com/majidbigdeli/neffosAmi/domin/data"
	"github.com/majidbigdeli/neffosAmi/domin/dto"
	"github.com/majidbigdeli/neffosAmi/domin/variable"
)

//BroadcastHandler ...
func BroadcastHandler(w http.ResponseWriter, r *http.Request) {

	var userMsg dto.FormDto

	if r.Body == nil {
		http.Error(w, "Please send a request body", 400)
		return
	}
	err := json.NewDecoder(r.Body).Decode(&userMsg)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	userID, err := data.GetUserIDByExtentionNumber(userMsg.Extension)

	if err != nil {
		http.Error(w, "userId not found", 500)
	}

	userNumber := strconv.Itoa(userID)

	nsConn.Conn.Server().Broadcast(nil, neffos.Message{
		To:        userNumber,
		Namespace: variable.Agent,
		Event:     variable.OnShowForm,
		Body:      neffos.Marshal(userMsg),
	})

	data.InsertLogForForm(userMsg.Extension, userMsg.Direction, 1, userMsg.CallID, userMsg.CallerNumber)

	w.WriteHeader(http.StatusOK)
}

//NotificationHandler ....
func NotificationHandler() {

	for k, v := range connections {
		if !v.IsClosed() {
			//delete(connections, k)
			id, _ := strconv.Atoi(k)
			notification, err := data.GetNotifByUserID(id)
			if err != nil {
				return
			}
			nsConn.Conn.Server().Broadcast(nil, neffos.Message{
				Namespace: variable.Agent,
				Event:     "resiveErja",
				To:        k,
				Body:      neffos.Marshal(notification),
			})
		}
	}

}
