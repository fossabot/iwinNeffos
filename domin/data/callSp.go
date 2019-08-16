package data

import (
	"database/sql"
	"fmt"

	mssql "github.com/denisenkom/go-mssqldb"
	"github.com/majidbigdeli/iwinNeffos/domin/model"
)

//GetNotificationList ...
//گرفتن لیست نوتیفیکیشن ها
// لیست اکستنشن ها را پاس می دهیم مدل IDTvp
// در نتیجه مدل Notification میگیریم
func GetNotificationList(listUserID []model.IDTvp) (*[]model.Notification, error) {

	if len(listUserID) > 0 {
		tvpType := mssql.TVP{
			TypeName: "typ.BigIntIDType",
			Value:    listUserID,
		}

		var d []model.Notification
		err := dbData.Select(&d, "[app].[uspGetNotificationListByUserID1]",
			sql.Named("PI_UserId", tvpType),
		)
		return &d, err
	}
	return nil, nil
}

//UpdateNotificationList ....
func UpdateNotificationList(listNotifID []model.IDTvp, status int) {

	tvpType := mssql.TVP{
		TypeName: "typ.BigIntIDType",
		Value:    listNotifID,
	}

	_, err := dbData.Exec("[app].[uspUpdateNotificationList1]",
		sql.Named("PI_NotificationId", tvpType),
		sql.Named("PI_Status", status),
	)
	if err != nil {
		fmt.Printf(err.Error())
	}

}

//UpdateNotification ...
func UpdateNotification(id, status int) {
	_, err := dbData.Exec("[app].[uspUpdateNotification1]",
		sql.Named("PI_NotificationId", id),
		sql.Named("PI_Status", status),
	)
	if err != nil {
		fmt.Printf(err.Error())
	}

}

//SetNeffosError ....
func SetNeffosError(neffosError model.NeffosError) {
	_, err := db.Exec("[logging].[UspSetNeffosError1]",
		sql.Named("PI_SocketId", neffosError.SocketID),
		sql.Named("PI_Message", neffosError.Message),
		sql.Named("PI_Body", neffosError.Body),
	)
	if err != nil {
		fmt.Printf(err.Error())
	}

}

//ChangeSendingStatusToNew ....
func ChangeSendingStatusToNew() {
	_, err := dbData.Exec("[app].[uspChangeSendingStatus1]")
	if err != nil {
		fmt.Printf(err.Error())
	}
}
