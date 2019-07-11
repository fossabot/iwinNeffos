package dto

//FormDto is dto
// مدلی که خانم وحید به من ارسال می کند
type FormDto struct {
	Extension    int   `json:"Extension"`
	Direction    int   `json:"Direction"`
	CallID       int64 `json:"CallId"`
	CallerNumber int64 `json:"CallerNumber"`
}
