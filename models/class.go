package model

// 班型
type Class struct {
	Id            int64    `json:"id"`
	CoachId       int64    `form:"coachId" json:"coachId"`
	Created       JsonTime `xorm:"created" json:"created"`
	SchoolName    string   `form:"schoolName" json:"schoolName"`
	DriveSchoolId int64    `form:"driveSchoolId" json:"driveSchoolId"`
	Name          string   `json:"name"`
	Type          int      `json:"type"` // 1－c1;2-c2
	Price         int      `json:"price"`
	Updated       JsonTime `xorm:"updated" json:"updated"`
}
