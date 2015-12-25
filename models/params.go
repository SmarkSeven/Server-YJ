package model

// 注册（或者重置密码）参数
type RegisterParams struct {
	Phone    string
	Password string
	Code     string
}

// 登录 参数
type LoginParams struct {
	Phone    string
	Password string
}

// 修改用户信息 参数
type ModifyCoachInfo struct {
	Name                 string
	Birthday             string
	Sex                  int // 1－男 2-女
	Introduction         string
	TeachSite            string `form:"teachSite"`
	TeachType            int    `form:"teachType"`
	DrivingSchool        string `form:"drivingSchool"`
	TeachCharacteristics string `form:"teachCharacteristics"` // 教学特色 “特色1-特色2”
	TeachSiteId          int    `form:"teachSiteId"`
	DrivingSchoolId      int    `form:"drivingSchoolId"`
}

// 添加学员 参数
type AddStudentsParams struct {
	Students string
}

//  获取学员列表 参数
type GetStudentsParams struct {
	Status int
}

// 修改学员备注
type ModifyRemrakParams struct {
	Id     int64
	Remark string
}

// 搜索学员 参数 每次只能上传一个参数
type SearchStudentParams struct {
	Phone string
	Name  string
}

// 修改学员的状态 参数
type ModifyStudentStatusParams struct {
	Id     int64
	Status int
}

// 图片回调 参数
type PictureCallBackParams struct {
	ImageType int `form:"imageType"`
}

// 删除图片 参数
type DeletePictureParams struct {
	Id int
}

// 搜索驾校 参数
type SearchSchoolParams struct {
	Name string
}

// 排程参数
type ScheduleParams struct {
	Day     int
	Hour    int
	Subject int
}
type ModifyScheduleParams struct {
	ScheduleId int64 `form:"scheduleId"`
	Subject    int
}
type BindingParams struct {
	StudentId int64 `form:"studentId"`
}
