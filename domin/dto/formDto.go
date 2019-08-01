package dto

import "encoding/json"

//FormDto is dto
// مدلی که خانم وحید به من ارسال می کند
type FormDto struct {
	Extension    int   `json:"Extension"`
	Direction    int   `json:"Direction"`
	CallID       int64 `json:"CallId"`
	CallerNumber int64 `json:"CallerNumber"`
}

//MarshalBinary ...
func (s FormDto) MarshalBinary() ([]byte, error) {
	return json.Marshal(s)
}

//UnmarshalBinary ...
func (s *FormDto) UnmarshalBinary(b []byte) error {
	return json.Unmarshal(b, &s)
}

//ExtensionModel ....
type ExtensionModel struct {
	Number int `json:"Number"`
}
