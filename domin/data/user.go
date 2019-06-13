package data

import (
	"database/sql"
	"fmt"
)

//InsertExtensionStatus ....
func InsertExtensionStatus(exten, status int) {
	_, err := db.Exec("service.uspUpdateAgentStatus1",
		sql.Named("PI_Number", exten),
		sql.Named("LastName", status),
	)
	if err != nil {
		fmt.Printf(err.Error())
	}
}
