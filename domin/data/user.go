package data

import (
	"database/sql"
	"fmt"
)

//InsertExtensionStatus ....
func InsertExtensionStatus(exten int, status int) {
	_, err := db.Exec("service.uspUpdateAgentStatus1",
		sql.Named("PI_Number", exten),
		sql.Named("PI_Status", status),
	)
	if err != nil {
		fmt.Printf(err.Error())
	}
}

//InsertAgentQueueStatus ....
func InsertAgentQueueStatus(extensionNumber, queueNumber, status int) {
	_, err := db.Exec("service.uspUpdateAgentQueueStatus2",
		sql.Named("PI_ExtensionNumber", extensionNumber),
		sql.Named("PI_ManagerQueueId", queueNumber),
		sql.Named("PI_Status", status),
	)
	if err != nil {
		fmt.Printf(err.Error())
	}
}
