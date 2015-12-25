// 教练自身相关
package router

import (
	"github.com/go-macaron/session"
	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
	"gopkg.in/macaron.v1"
	"strings"
	"yj-server-golang/models"
)

// 修改信息
func ModifyInfo(c *macaron.Context, params model.ModifyCoachInfo, sess session.Store) {
	hasParams := false
	id := getCoachId(c, sess)
	if id == 0 {
		return
	}
	coach := model.Coach{Id: id}
	ok, err := engine.Get(coach)
	if err != nil {
		LLog("dengine.Get(coach) in ModifyInfo()", err)
		Errors(c)
		return
	}
	if len(params.Name) > 0 {
		coach.Name = params.Name
		hasParams = true
	}
	if len(params.Birthday) > 0 {
		coach.Birthday = params.Birthday
		hasParams = true
	}
	if len(params.Introduction) > 0 {
		coach.Introduction = params.Introduction
		hasParams = true
	}
	if len(params.TeachSite) > 0 {
		coach.TeachSite = params.TeachSite
		hasParams = true
	}
	if len(params.DrivingSchool) > 0 {
		coach.DrivingSchool = params.DrivingSchool
		hasParams = true
	}
	if len(params.TeachCharacteristics) > 0 {
		coach.TeachCharacteristics = params.TeachCharacteristics
		hasParams = true
	}

	if params.Sex != 0 {
		if params.Sex == 1 || params.Sex == 2 {
			coach.Sex = params.Sex
			hasParams = true
		} else {
			c.JSON(200, RespError{false, "参数sex错误"})
		}
	}
	if params.TeachType != 0 {
		if params.TeachType == 1 || params.TeachType == 2 {
			coach.TeachType = params.TeachType
			hasParams = true
		} else {
			c.JSON(200, RespError{false, "参数teachType错误"})
		}
	}
	if hasParams {
		// 开启事务
		session := engine.NewSession()
		err = session.Begin()
		defer session.Close()
		if err != nil {
			LLog("session.Begin() in ModifyInfo", err)
			Errors(c)
			return
		}

		// 更新信息
		num, err := session.Id(coach.Id).Update(&coach)
		if err != nil {
			LLog("Update(&coach) in ModifyInfo", err)
			Errors(c)
			session.Rollback()
			return
		}
		if num == 0 {
			c.JSON(200, RespError{false, "未做任何修改"})
			session.Rollback()
			return
		}
		err = session.Commit()
		if err != nil {
			LLog("session.Commit() in ModifyInfo", err)
			c.JSON(200, RespError{false, "重置密码失败"})
			session.Rollback()
			return
		}
		engine.Get(&coach)
		if err != nil {
			Errors(c)
			return
		}
		if ok {
			sess.Set("coach", coach)
		}
		coach.Password = ""
		c.JSON(200, RespInfo{true, coach})
		return
	}
	c.JSON(200, RespError{false, "未提交参数"})
}

// 提交审核申请
func ApplyToCoach(c *macaron.Context, sess session.Store) {
	coachId := getCoachId(c, sess)
	if coachId == 0 {
		return
	}
	coach := model.Coach{Id: coachId}
	ok, err := engine.Get(&coach)
	if err != nil {
		LLog("engine.Get(&coach) in ApplyToCoach", err)
		Errors(c)
		return
	}
	if !ok {
		c.JSON(200, RespError{false, "获取教练信息失败"})
		return
	}
	if len(coach.Identity) <= 0 {
		c.JSON(200, RespError{false, "您还未上传身份证照片"})
		return
	}
	if len(coach.DriveLisence) <= 0 {
		c.JSON(200, RespError{false, "您还未上传驶证照片"})
		return
	}
	if len(coach.CoachLisence) <= 0 {
		c.JSON(200, RespError{false, "您还未上传教练证照片"})
		return
	}
	if coach.Status == 2 {
		ErrorJson(c, "您已经通过审核")
		return
	}
	if coach.Status == 3 {
		ErrorJson(c, "您已经提交过申请，申请正在审核中")
		return
	}
	if coach.Status == 4 {
		ErrorJson(c, "您上次的审核未通过,请按提示操作后再申请")
		return
	}
	kvs := make(map[string]interface{})
	kvs["coach"] = coach
	commitWithDB(c, kvs, func(c *macaron.Context, kvs map[string]interface{}, s *xorm.Session) interface{} {
		v, _ := kvs["coach"]
		coach := v.(model.Coach)
		coach.Status = 3
		_, err := s.Id(coach.Id).Cols("Status").Update(&coach)
		if err != nil {
			LLog("s.Id(coach.Id).Cols(Status).Update(&coach) in ApplyToCoach", err)
			Errors(c)
			return nil
		}
		return RespInfo{true, coach}
	})
}

// 创建班型
func NewClass(c *macaron.Context, class model.Class, sess session.Store) {
	if class.DriveSchoolId <= 0 {
		c.JSON(200, RespError{false, "参数driveSchoolId不能为空"})
		return
	}
	// // 参数验证
	// if strings.EqualFold(class.ShoolName, "") {
	// 	c.JSON(200, RespError{false, "'所属驾校'不能为空"})
	// 	return
	// }

	school := model.DriveSchool{Id: class.DriveSchoolId}
	ok, err := engine.Get(&school)
	if err != nil {
		LLog("dengine.Get(&school) in ModifyCalss()", err)
		Errors(c)
		return
	}
	// if !ok {
	// 	c.JSON(200, RespError{false, "参数driveSchoolId和shoolName不匹配"})
	// 	return
	// }
	if !ok {
		c.JSON(200, RespError{false, "您选择的驾校不存在"})
		return
	}
	class.SchoolName = school.Name
	// 查看数据库中是否有此驾校
	if strings.EqualFold(class.Name, "") {
		c.JSON(200, RespError{false, "'班级名称'不能为空"})
		return
	}

	if class.Type < 1 || class.Type > 2 {
		c.JSON(200, RespError{false, "参数type错误"})
		return
	}
	if class.Price <= 0 {
		c.JSON(200, RespError{false, "参数Price错误"})
		return
	}

	// 获取教练Id
	coachId := getCoachId(c, sess)
	if coachId == 0 {
		return
	}
	// 检查是否已存在
	class.CoachId = coachId

	ok, err = engine.Get(&class)
	if err != nil {
		LLog("engine.Get(&class)  in newClass", err)
		Errors(c)
		return
	}
	if ok {
		c.JSON(200, RespError{false, "只能创建一个班型"})
		return
	}

	// 处理数据
	kvs := make(map[string]interface{})
	kvs["class"] = class
	commitWithDB(c, kvs, func(c *macaron.Context, kvs map[string]interface{}, s *xorm.Session) interface{} {
		v, _ := kvs["class"]
		class := v.(model.Class)
		// commitWithDB(c, class, func(c *macaron.Context, v interface{}, s *xorm.Session) interface{} {
		// 	class := v.(model.Class)
		//  添加班型
		num, err := s.Insert(&class)
		if err != nil {
			LLog("s.Insert(&class)  in newClass", err)
			Errors(c)
			return nil
		}
		if num == 0 {
			c.JSON(200, RespError{false, "创建班型失败"})
			return nil
		}
		// 更新Coach
		coach := model.Coach{Id: class.CoachId, Price: class.Price, DrivingSchool: class.SchoolName}
		num, err = s.Id(coach.Id).Update(&coach)
		if err != nil {
			LLog("s.Id(coach.Id).Cols(Price).Update(&coach) in newClass", err)
			Errors(c)
			return nil
		}
		if num == 0 {
			c.JSON(200, RespError{false, "创建班型失败"})
			return nil
		}
		return RespData{true, class}
	})
}

// 修改班型
func ModifyCalss(c *macaron.Context, class model.Class, sess session.Store) {
	// 参数验证
	var hasParams bool
	if class.DriveSchoolId > 0 {
		school := model.DriveSchool{Id: class.DriveSchoolId}
		ok, err := engine.Get(&school)
		if err != nil {
			LLog("dengine.Get(&school) in ModifyCalss()", err)
			Errors(c)
			return
		}
		if !ok {
			c.JSON(200, RespError{false, "您选择的加驾校不存在"})
			return
		}
		class.SchoolName = school.Name
		hasParams = true
	}
	// if !strings.EqualFold(class.SchoolName, "") {
	// 	hasParams = true
	// }
	// 查看数据库中是否有此驾校
	if !strings.EqualFold(class.Name, "") {
		hasParams = true
	}
	if class.Type >= 1 && class.Type <= 2 {
		hasParams = true
	}
	if class.Price > 0 {
		hasParams = true
	}

	if hasParams {
		// 获取教练Id
		coachId := getCoachId(c, sess)
		if coachId == 0 {
			return
		}
		class.CoachId = coachId
		// 查找班型
		temp := model.Class{CoachId: coachId}
		ok, err := engine.Get(&temp)
		if err != nil {
			LLog("engine.Get(&temp) in newClass", err)
			Errors(c)
			return
		}
		if !ok {
			c.JSON(200, RespError{false, "您还没有创建班型"})
			return
		}
		class.Id = temp.Id
		// 处理数据
		kvs := make(map[string]interface{})
		kvs["class"] = class
		commitWithDB(c, kvs, func(c *macaron.Context, kvs map[string]interface{}, s *xorm.Session) interface{} {
			v, _ := kvs["class"]
			class := v.(model.Class)
			// commitWithDB(c, class, func(c *macaron.Context, v interface{}, s *xorm.Session) interface{} {
			// 	class := v.(model.Class)
			num, err := s.Id(class.Id).Update(&class)
			if err != nil {
				LLog("s.Id(class.Id).Update(&class) in newClass", err)
				Errors(c)
				return nil
			}
			if num == 0 {
				c.JSON(200, RespError{false, "未做任何修改"})
				return nil
			}
			return RespData{true, class}
		})
		return
	}
	c.JSON(200, RespError{false, "没有参数"})
}

// 获取版型信息
func GetClassInfo(c *macaron.Context, sess session.Store) {
	coachId := getCoachId(c, sess)
	class := model.Class{CoachId: coachId}
	ok, err := engine.Get(&class)
	if err != nil {
		LLog("engine.Get(&class) in GetClassInfo()", err)
		Errors(c)
		return
	}
	if !ok {
		c.JSON(200, RespSuccese{false})
		return
	}
	c.JSON(200, RespData{true, class})
}

type RespSchools struct {
	Success bool                `json:"success"`
	Schools []model.DriveSchool `json:"schools"`
}

// 搜索驾校
func SearchDriveSchool(c *macaron.Context, sess session.Store) {
	c.Req.ParseForm()
	name := c.Query("name")
	if len(name) == 0 {
		LLog("NAME == 0", nil)
		schools := make([]model.DriveSchool, 0)
		engine.Find(&schools)
		c.JSON(200, RespSchools{true, schools})
		return
	}
	schools := make([]model.DriveSchool, 0)
	name = "%" + name + "%"
	engine.Sql("select * from driveSchool where name like ?", name).Find(&schools)
	c.JSON(200, RespSchools{true, schools})
	return
}
