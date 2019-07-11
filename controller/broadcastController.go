package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/kataras/neffos"
	"github.com/majidbigdeli/neffosAmi/domin/data"
	"github.com/majidbigdeli/neffosAmi/domin/dto"
	"github.com/majidbigdeli/neffosAmi/domin/model"
	"github.com/majidbigdeli/neffosAmi/domin/variable"
)

var mutex = &sync.Mutex{}

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

	extensionNumber := strconv.Itoa(userMsg.Extension)

	Server.Broadcast(nil, neffos.Message{
		To:        extensionNumber,
		Namespace: variable.Agent,
		Event:     variable.OnShowForm,
		Body:      neffos.Marshal(userMsg),
	})

	data.InsertLogForForm(userMsg.Extension, userMsg.Direction, 1, userMsg.CallID, userMsg.CallerNumber)

	w.WriteHeader(http.StatusOK)
}

//NotificationHandler ...
func NotificationHandler() {
	mutex.Lock()

	//listExtensionNumber , connectionID = extentionNumber
	var listExtensionNumber []model.IDTvp

	connections := Server.GetConnections()
	for c := range connections {
		var extentionNumber model.IDTvp
		connectionID, err := strconv.Atoi(c)
		if err != nil {
			fmt.Println(err)
		}
		extentionNumber.ID = connectionID
		listExtensionNumber = append(listExtensionNumber, extentionNumber)
	}

	//get list Notification from database with send list ExtensionNumber ;
	// Attention { ExtensionNumber == list ConnectionID }
	listNotification, err := data.GetNotificationList(listExtensionNumber)

	if err != nil {
		fmt.Println(err)
		return
	}

	if listNotification != nil {
		if len(*listNotification) > 0 {
			for _, element := range *listNotification {
				//extensionNumber = my unique connection ID
				extensionNumber := strconv.Itoa(element.Number)
				Server.Broadcast(nil, neffos.Message{
					To:        extensionNumber,
					Namespace: variable.Agent,
					Event:     variable.OnNotification,
					Body:      neffos.Marshal(element),
				})
			}

		}
	}

	mutex.Unlock()

}
