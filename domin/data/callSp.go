package data

import (
	"database/sql"
	"fmt"

	mssql "github.com/denisenkom/go-mssqldb"
	"github.com/majidbigdeli/neffosAmi/domin/model"
)

//InsertLogForForm ...
// لاگ اول برای زمانی که خانم وحید ای پی آی را صدا میزند
func InsertLogForForm(extension, direction, typeSend int, callID, callerNumber int64) {
	_, err := db.Exec("logging.uspInsertLogForForm1",
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

//GetNotificationList ...
//گرفتن لیست نوتیفیکیشن ها
// لیست اکستنشن ها را پاس می دهیم مدل IDTvp
// در نتیجه مدل Notification میگیریم
func GetNotificationList(listExtentions []model.IDTvp) (*[]model.Notification, error) {
	var poError int32
	var poSTEP int32

	if len(listExtentions) > 0 {
		tvpType := mssql.TVP{
			TypeName: "typ.BigIntIDType",
			Value:    listExtentions,
		}

		var d []model.Notification
		err := dbData.Select(&d, "app.uspGetNotificationList1",
			sql.Named("PI_Extensions", tvpType),
			sql.Named("PO_Error", sql.Out{Dest: &poError}),
			sql.Named("PO_STEP", sql.Out{Dest: &poSTEP}),
		)
		return &d, err
	}
	return nil, nil
}

//GetExtentinByUserID ...
func GetExtentinByUserID(userID int) (int, error) {
	var extensionID *int

	err := dbCore.Get(&extensionID, "SELECT ExtensionID FROM acc.Agent WHERE UserID = @UserId", sql.Named("UserId", userID))

	if err != nil {
		return 0, err
	}

	return *extensionID, nil

}

//GetUserIDByExtention ...
func GetUserIDByExtention(extension string) (string, error) {
	var userID string

	err := dbCore.Get(&userID, "SELECT a.UserID FROM acc.Agent a WITH(NOLOCK) INNER JOIN (SELECT ExtensionID ,Number FROM  core.Extension WITH(NOLOCK)) e ON e.ExtensionID = a.ExtensionID WHERE e.Number = @Extention", sql.Named("Extention", extension))

	if err != nil {
		return "", err
	}

	return userID, nil

}

//GetUserIDByExtentionNumber ...
func GetUserIDByExtentionNumber(extension int) (string, error) {
	var userID string

	err := dbCore.QueryRowx("[app].[uspGetUserIdByExtension1]", sql.Named("PI_ExtensionNumber", extension)).Scan(&userID)

	if err != nil {
		return "", err
	}

	return userID, nil

}

//GetNotifByUserID ...
func GetNotifByUserID(userID int) (*[]model.Notification, error) {
	var d []model.Notification
	err := dbData.Select(&d, "SELECT  Data,	Message,MessageType,NotificationId,Status,Type,UserId FROM message.Notification WHERE [UserId] = @UserId AND [Status] = @Status",
		sql.Named("UserId", userID),
		sql.Named("Status", 22710),
	)
	return &d, err
}

//GetNotif ...
func GetNotif() (*[]model.Notification, error) {

	var d []model.Notification
	err := dbData.Select(&d, "SELECT  Data,	Message,MessageType,NotificationId,Status,Type,UserId FROM message.Notification WHERE [Status] = @Status",
		sql.Named("Status", 22710),
	)

	return &d, err
}

//GetExtentionUser ...
func GetExtentionUser() ([]model.ExtUser, error) {
	var extUser []model.ExtUser
	err := dbCore.Select(&extUser, "[app].[UspGetExtentionUser1]")
	return extUser, err

}

//UpdateNotification ...
func UpdateNotification(id int) {
	var poError int32
	var poSTEP int32
	_, err := dbData.Exec("app.uspUpdateNotification1",
		sql.Named("PI_NotificationId", id),
		sql.Named("PI_Status", 22712),
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

//SetNeffosError1 ....
func SetNeffosError1(neffosError model.NeffosError) {
	_, err := db.Exec("logging.UspSetNeffosError1",
		sql.Named("PI_SocketId", neffosError.SocketID),
		sql.Named("PI_Message", neffosError.Message),
		sql.Named("PI_Body", neffosError.Body),
	)

	if err != nil {
		fmt.Printf(err.Error())
	}

}

//GetNotificationTime ....
func GetNotificationTime() (string, error) {
	var notiftime string

	err := dbCore.QueryRowx("SELECT [Value] FROM ref.ConfigurationSetting WHERE [Key] = @Key",
		sql.Named("Key", "NotificationTime")).Scan(&notiftime)

	if err != nil {
		if err != sql.ErrNoRows {
			fmt.Println(err)
			return "1m", err
		}
		return "1m", fmt.Errorf("not found Key NotificationTime in Db")
	}

	return notiftime, nil
}
