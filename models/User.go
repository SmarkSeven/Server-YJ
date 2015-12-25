package model

type User struct {
	Id       int64
	RoleId   int64 `json:"roleId"`
	Phone    string
	Name     string
	Created  JsonTime `xorm:"created" json:"created"`
	Password string
	Role     int      // 1-教练 2-学员
	Updated  JsonTime `xorm:"updated" json:"updated"`
}

type Coach struct {
	Id                   int64    `json:"id"`
	Userid               int64    `json:"userid"`
	Phone                string   `json:"phone"`
	Password             string   `json:"password"`
	Created              JsonTime `xorm:"created" json:"Created"`
	Name                 string   `json:"name"`
	Birthday             string   `json:"birthday"`
	Sex                  int      `json:"sex"` // 1－男 2-女
	Introduction         string   `json:"introduction"`
	AutoSchedule         int      `json:"autoSchedule"` // 自动排课，默认值0，1-不自动排课
	TeachSite            string   `json:"teachSite"`
	TeachSiteId          int64    `json:"teachSiteId"`
	TeachType            int      `json:"teachType"`
	DrivingSchool        string   `json:"drivingSchool"`
	DrivingSchoolId      int64    `json:"drivingSchoolId"`
	TeachCharacteristics string   `json:"teachCharacteristics"` // 教学特色 “特色1-特色2”
	Avator               string   `json:"avator"`
	Identity             string   `json:"identity"`
	CoachLisence         string   `json:"coachLisence"`
	DriveLisence         string   `json:"driveLisence"`
	Status               int      `json:"status"`
	Remark               string   `json:"remark"`
	Price                int      `json:"price"`
	Updated              JsonTime `xorm:"updated" json:"updated"`
}

type Student struct {
	Id         int64    `json:"id"`
	CoachId    int64    `json:"coachId"`
	Sex        int      `json:"sex"`
	Sname      string   `json:"sname"`
	Telephone  string   `json:"telephone"`
	Created    JsonTime `xorm:"created" json:"created"`
	Remark     string   `json:"remark"`
	Sstatus    int      `json:"sstatus"`
	Official   int      `json:"official"` // 1－ 只是被被教练添加进来的学员，学员自身还没进入过平台；2-学员进入平台过
	Openid     string   `json:"openid"`
	Updated    JsonTime `xorm:"updated" json:"updated"`
	Avator     string   `xorm:"headimg" json:"avator"`
	Bindstatus int      `json:"bindstatus"`
}
