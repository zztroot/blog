package models

type Gzh struct {
	Model
	AccessToken string `json:"access_token" gorm:"text"`
	ExpiresIn uint `json:"expires_in"`
}


