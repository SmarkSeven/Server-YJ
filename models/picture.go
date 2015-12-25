package model

type Picture struct {
	Id        int64    `json:"id"`
	UserId    int64    `json:"userId"`
	Url       string   `json:"url"`
	Created   JsonTime `xorm:"created" json:"created"`
	ImageType int      `json:"imageType"`
}
