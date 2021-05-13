package models

func GetObjects() []interface{} {
	return []interface{}{Article{}, User{}, Type{}, ArticleComment{}, WebComment{}, ImgOfArticle{}, Info{}, Md5Info{}, Gzh{}}
}

type Model struct {
	Id        uint   `gorm:"primary_key" json:"id"`
	CreatedAt string `json:"created_time"`
	UpdatedAt string `json:"updated_time"`
	DeleteAt  string `json:"delete_at"` //为no证明没有被删除/yes为删除
}
