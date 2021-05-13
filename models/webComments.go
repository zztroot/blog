package models

//网站留言
type WebComment struct {
	Model
	UserId     uint   `json:"user_id"`
	TempUserName string   `json:"temp_user_name"`
	Content    string `json:"content" gorm:"text"`
	ToUserId   uint   `json:"to_user_id"`
	ToTempUserName string `json:"to_temp_user_name"`
	LikeCount  uint   `json:"like_count"`
}