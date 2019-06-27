package data

import (
	"database/sql"
	"fmt"
)

//InsertExtensionStatus ....
func InsertExtensionStatus(exten int, status int) {
	var poError int32
	var poSTEP int32
	_, err := db.Exec("service.uspUpdateAgentStatus1",
		sql.Named("PI_Number", exten),
		sql.Named("PI_Status", status),
		sql.Named("PO_Error", sql.Out{Dest: &poError}),
		sql.Named("PO_STEP", sql.Out{Dest: &poSTEP}),
	)

	if poError > 0 {
		fmt.Printf("Error in db : [%d]", poError)
	}

	if err != nil {
		fmt.Printf(err.Error())
	}
}

//InsertAgentQueueStatus ....
func InsertAgentQueueStatus(extensionNumber, queueNumber, status int) {
	var poError int32
	var poSTEP int32

	_, err := db.Exec("service.uspUpdateAgentQueueStatus2",
		sql.Named("PI_ExtensionNumber", extensionNumber),
		sql.Named("PI_ManagerQueueId", queueNumber),
		sql.Named("PI_Status", status),
		sql.Named("PO_Error", sql.Out{Dest: &poError}),
		sql.Named("PO_STEP", sql.Out{Dest: &poSTEP}),
	)

	if poError > 0 {
		fmt.Printf("Error in db : [%d]", poError)
	}

	if err != nil {
		fmt.Printf(err.Error())
	}
}

//InsertLogForForm ...
func InsertLogForForm(extension, direction, typeSend int, callID, callerNumber int64) {
	_, err := db.Exec("service.uspInsertLogForForm1",
		sql.Named("PI_Extension", extension),
		sql.Named("PI_Direction", direction),
		sql.Named("PI_CallID", callID),
		sql.Named("PI_CallerNumber", callerNumber),
		sql.Named("PI_Type", typeSend),
	)
	if err != nil {
		fmt.Printf(err.Error())
	}
}