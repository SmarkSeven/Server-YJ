package model

import (
// "time"
)

// 排程
type Schedule struct {
	Id       int64    `json:"id"`
	CoachId  int64    `json:"coachId"`
	Number   int      `json:"number"`   // 当前数量
	Max      int      `json:"max"`      // 最大数量
	Subject  int      `json:"subject"`  // 科目 1-科目一，2-科目二，3-科目三
	Datetime JsonTime `json:"datetime"` // 排程的时间
	Status   int      `json:"status"`   // 默认值为1－表示可以更改［变更科目、取消排课］；2-表示有人预约，不可更改；3-表示被取消的排程
	Source   int      // 默认值0，1-来自于自动排课
	Created  JsonTime `xorm:"created" json:"created"`
	Updated  JsonTime `xorm:"updated" json:"updated"`
}

// // 自动排程时间表
// type AutoSchedule struct {
// 	Id      int64 `json:"id"`
// 	CoachId int64 `json:"coachId"`
// 	// Datetime string   `json:"datetime"` // 自动排课的课程时间
// 	Created JsonTime `xorm:"created" json:"created"`
// 	Updated JsonTime `xorm:"updated" json:"updated"`
// }
