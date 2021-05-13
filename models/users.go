package models

//accounts
type User struct {
	Model
	UserName string `json:"username"`
	RealName string `json:"real_name"`
	Pwd      string `json:"pwd"`
	Age      uint   `json:"age"`
	Sex      string `json:"sex"`
	Address  string `json:"address"`
	Phone    string `json:"phone"`
}
