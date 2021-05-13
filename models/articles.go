package models

//文章
type Article struct {
	Model
	ArticleName    string `json:"article_name"`
	ArticleContent string `json:"article_content" gorm:"text"`
	TypesId        uint   `json:"types_id"`
	ReadCount      uint   `json:"read_count" gorm:"-"`
	LikeCount      uint   `json:"like_count" gorm:"-"`
	CommentCount   uint   `json:"comment_count" gorm:"-"`
	UserId         uint   `json:"user_id"`
	ImgId          uint   `json:"img_id"`
}

//文章类型
type Type struct {
	Model
	TypeName string `json:"type_name"`
}

//文章留言
type ArticleComment struct {
	Model
	UserId         uint   `json:"user_id"`
	TempUserName   string `json:"temp_user_name"`
	ArticleId      uint   `json:"article_id"`
	Content        string `json:"content" gorm:"text"`
	ToUserId       uint   `json:"to_user_id"`
	ToTempUserName string `json:"to_temp_user_name"`
	LikeCount      uint   `json:"like_count"`
}

//文章图片
type ImgOfArticle struct {
	Model
	ImgPath   string `json:"img_path"`
	ImgIndex  uint   `json:"img_index"`
}

type Info struct {
	Model
	Content string `json:"content"`
}
