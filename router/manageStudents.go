package router

import (
	"encoding/json"
	"github.com/go-macaron/session"
	"github.com/go-xorm/xorm"
	"gopkg.in/macaron.v1"
	"log"
	"strings"
	"yj-server-golang/models"
)

type StudentsParams struct {
	Students []model.Student
}

type RespStudnets struct {
	Success  bool            `json:"success"`
	Students []model.Student `json:"students"`
}

type RespStudent struct {
	Success bool          `json:"success"`
	Student model.Student `json:"student"`
}

// 添加学员
func AddStudents(c *macaron.Context, param model.AddStudentsParams, sess session.Store) {
	if strings.EqualFold(param.Students, "") {
		c.JSON(200, RespError{false, "参数为空"})
		return
	}
	studentsParams := StudentsParams{}
	err := json.Unmarshal([]byte(param.Students), &studentsParams)
	if err != nil {
		LLog("err := json.Unmarshal([]byte(param.Students), &studentsParams) in AddStudents", err)
		c.JSON(200, RespError{false, "参数格式错误"})
		return
	}
	if len(studentsParams.Students) == 0 {
		c.JSON(200, RespError{false, "没有上传学员信息"})
		return
	}
	// 获取教练Id
	id := getCoachId(c, sess)
	if id == 0 {
		return
	}
	// c.Data["coachId"] = id
	kvs := make(map[string]interface{})
	kvs["studentsParams"] = studentsParams
	kvs["coachId"] = id
	commitWithDB2(c, kvs, func(c *macaron.Context, kvs map[string]interface{}, s *xorm.Session) bool {
		v, _ := kvs["studentsParams"]
		studentsParams := v.(StudentsParams)
		v2, _ := kvs["coachId"]
		coachId := v2.(int64)
		// commitWithDB2(c, studentsParams, func(c *macaron.Context, v interface{}, s *xorm.Session) bool {
		// 	studentsParams := v.(StudentsParams)
		for _, paramStudent := range studentsParams.Students {
			log.Println("for _, paramStudent := range studentsParams.Students")
			if len(paramStudent.Sname) == 0 {
				c.JSON(200, RespError{false, "请添加学员的名字"})
				return false
			}
			if len(paramStudent.Telephone) == 0 {
				c.JSON(200, RespError{false, "请添加学员的手机号码"})
				return false
			}
			if paramStudent.Sstatus != 1 && paramStudent.Sstatus != 2 {
				c.JSON(200, RespError{false, "请添加学员的的学习状态"})
				return false
			}
			if !CheckPhone(paramStudent.Telephone) {

				c.JSON(200, RespError{false, "手机号码格式错误"})
				return false
			}
			ok, err := engine.Get(&model.Coach{Phone: paramStudent.Telephone})
			if ok {
				errorInfo := "手机号:" + paramStudent.Telephone + "已被注册为教练"
				ErrorJson(c, errorInfo)
				return false
			}
			student := model.Student{Telephone: paramStudent.Telephone}
			has, err := engine.Get(&student)
			if err != nil {
				LLog("engine.Get(&student)in AddStudents()", err)
				Errors(c)
				return false
			}
			// id = c.Data["coachId"].(int64)
			if has && student.CoachId > 0 {
				if id != student.CoachId {
					c.JSON(200, RespError{false, "学员:" + paramStudent.Sname + "(" + paramStudent.Telephone + ")" + "已经被其他教练绑定"})
					return false
				}
				c.JSON(200, RespError{false, "您已经绑定过学员:" + paramStudent.Sname + "(" + paramStudent.Telephone + ")"})
				return false
			}
			if student.Official == 0 {
				paramStudent.Official = 1
			}
			paramStudent.CoachId = coachId
			if has {
				num, err := s.Id(student.Id).Update(&paramStudent)
				if err != nil {
					LLog("num, err := s.Update(&paramStudent) in AddStudents()", err)
					c.JSON(200, RespError{false, "系统错误"})
					return false
				}
				if num == 0 {
					c.JSON(200, RespError{false, "添加学员失败"})
					return false
				}
				return true
			}
			num, err := s.Insert(&paramStudent)
			if err != nil {
				Errors(c)
				return false
			}
			if num == 0 {
				c.JSON(200, RespError{false, "添加学员失败"})
				return false
			}
		}
		return true
	}, func(c *macaron.Context, kvs map[string]interface{}, s *xorm.Session) interface{} {
		// <<<<<<< HEAD
		// 		v, _ := kvs["coachId"]
		// 		id := v.(int64)
		// =======
		// 		// id := c.Data["coachId"].(int64)
		// >>>>>>> https
		//获取所有学员
		v, _ := kvs["coachId"]
		coachId := v.(int64)
		student := model.Student{CoachId: coachId}
		students := make([]model.Student, 0)
		err := engine.Find(&students, student)
		if err != nil {
			Errors(c)
			return nil
		}
		return RespStudnets{true, students}
	})
}

// 删除学员
func DeleteStudent(c *macaron.Context, sess session.Store) {
	// 1-检查参数，2-检查是否有该记录，3-数据处理
	studentId := c.ParamsInt64("id")
	if studentId <= 0 {
		resultError(c, 3, "无效的参数ID")
		return
	}
	coachId := getCoachId(c, sess)
	if coachId == 0 {
		resultError(c, 6, "")
		return
	}
	student := model.Student{Id: studentId, CoachId: coachId}
	ok, err := engine.Get(&student)
	if err != nil {
		LLog("ok, err := engine.Get(&student)in DeleteStudent", err)
		resultError(c, 6, "")
		return
	}
	if !ok {
		resultError(c, 7, "没有此学员,不能删除")
		return
	}

	student.CoachId = -1
	// 处理数据
	kvs := make(map[string]interface{})
	kvs["student"] = student
	commitWithDB(c, kvs, func(c *macaron.Context, kvs map[string]interface{}, s *xorm.Session) interface{} {
		v, _ := kvs["student"]
		student := v.(model.Student)
		// commitWithDB(c, picture, func(c *macaron.Context, v interface{}, s *xorm.Session) interface{} {
		// 	picture := v.(model.Picture)
		// num, err := s.Delete(&student)
		num, err := s.Id(student.Id).Cols("CoachId").Update(&student)
		if err != nil {
			LLog("s.Delete(&picture) in DeletePicture", err)
			resultError(c, 6, "")
			return nil
		}
		if num == 0 {
			resultError(c, 7, "删除图片失败")
			return nil
		}
		return model.Result{Code: 1}
	})

}

// 获取学员列表
func GetStudents(c *macaron.Context, param model.GetStudentsParams, sess session.Store) {
	if param.Status < 1 && param.Status > 2 {
		c.JSON(200, RespError{false, "参数错误"})
		return
	}
	// 获取教练Id
	kvs := make(map[string]interface{})
	if sess.Get("coach") == nil {
		c.JSON(200, RespError{true, "没有登录"})
		return
	}
	kvs = map[string]interface{}(sess.Get("coach").(map[string]interface{}))
	if kvs == nil {
		c.JSON(200, RespError{false, "没有登录"})
		sess.Destory(c)
		return
	}
	id := int64(kvs["id"].(float64))
	//按状态获取学员列表
	student := model.Student{CoachId: id, Sstatus: param.Status}
	students := make([]model.Student, 0)
	err := engine.Find(&students, student)
	if err != nil {
		LLog("engine.Find(&students, student) in GetStudents", err)
		Errors(c)
		return
	}
	c.JSON(200, RespStudnets{true, students})
}

// 获取学员信息 参数学员ID
func GetStudentInfo(c *macaron.Context, sess session.Store) {
	studentId := c.QueryInt64("id")
	if studentId == 0 {
		c.JSON(200, RespError{false, "参数错误"})
		return
	}
	student := model.Student{Id: studentId}
	ok, err := engine.Get(&student)
	if err != nil {
		LLog("engine.Get(&student) in GetStudentInfo()", err)
		Errors(c)
		return
	}
	if !ok {
		c.JSON(200, RespError{false, "不存在此ID"})
		return
	}
	// 获取教练Id
	kvs := make(map[string]interface{})
	if sess.Get("coach") == nil {
		c.JSON(200, RespError{true, "没有登录"})
		return
	}
	kvs = map[string]interface{}(sess.Get("coach").(map[string]interface{}))
	if kvs == nil {
		c.JSON(200, RespError{false, "没有登录"})
		return
	}
	id := int64(kvs["id"].(float64))
	if id != student.CoachId {
		c.JSON(200, RespError{false, "不能获取非自己学员的信息"})
		return
	}
	c.JSON(200, RespStudent{true, student})
}

// 修改学员备注
func ModifyRemrak(c *macaron.Context, params model.ModifyRemrakParams, sess session.Store) {
	// 参数验证
	if params.Id == 0 {
		c.JSON(200, RespError{true, "id不能为0"})
		return
	}
	if len(params.Remark) == 0 {
		c.JSON(200, RespError{true, "remark不能为空"})
		return
	}
	// 获取教练Id
	id := getCoachId(c, sess)
	if id == 0 {
		return
	}
	// 查找学员
	student := model.Student{Id: params.Id}
	ok, err := engine.Get(&student)
	if err != nil {
		LLog("engine.Get(&student) in ModifyRemrak()", err)
		Errors(c)
		return
	}
	if !ok {
		c.JSON(200, RespError{false, "不存在此ID"})
		return
	}
	if id != student.CoachId {
		c.JSON(200, RespError{false, "不能修改非自己学员的备注信息"})
		return
	}
	student.Remark = params.Remark
	// 数据处理
	kvs := make(map[string]interface{})
	kvs["student"] = student
	commitWithDB(c, kvs, func(c *macaron.Context, kvs map[string]interface{}, s *xorm.Session) interface{} {
		v, _ := kvs["student"]
		student := v.(model.Student)
		// commitWithDB(c, student, func(c *macaron.Context, v interface{}, s *xorm.Session) interface{} {
		// 	student := v.(model.Student)
		num, err := s.Id(student.Id).Cols("Remark").Update(&student)
		if err != nil {
			LLog("s.Id(student.Id).Cols(Remark).Update(&student) in ModifyRemrak()", err)
			Errors(c)
			return nil
		}
		if num == 0 {
			c.JSON(200, RespError{false, "未做任何修改"})
			return nil
		}
		return RespStudent{true, student}
	})
}

// 获取所有学员
func GetAllStudents(c *macaron.Context, sess session.Store) {
	// 教练ID
	coachId := getCoachId(c, sess)
	if coachId == 0 {
		return
	}
	student := model.Student{CoachId: coachId}
	students := make([]model.Student, 0)
	err := engine.Find(&students, &student)
	if err != nil {
		LLog("engine.Find(&students, &student) in ModifyRemrak", err)
		Errors(c)
		return
	}
	c.JSON(200, RespStudnets{true, students})
}

// 根据添条件搜索学员
func SearchStudent(c *macaron.Context, params model.SearchStudentParams, sess session.Store) {
	// 验证参数
	if len(params.Phone) > 0 && len(params.Name) > 0 {
		c.JSON(200, RespError{false, "只能有一个查询条件"})
		return
	}
	if len(params.Phone) == 0 && len(params.Name) == 0 {
		c.JSON(200, RespError{false, "查询条件有误"})
		return
	}
	id := getCoachId(c, sess)
	if id == 0 {
		return
	}
	student := model.Student{CoachId: id}
	if len(params.Phone) > 0 {
		student.Telephone = params.Phone
	}
	if len(params.Name) > 0 {
		student.Sname = params.Name
	}

	students := make([]model.Student, 0)
	err := engine.Find(&students, &student)
	if err != nil {
		LLog("engine.Find(&students, &student) in SearchStudent()", err)
		Errors(c)
		return
	}
	c.JSON(200, RespStudnets{true, students})
}

//  修改学员的状态
func ModifyStudentStatus(c *macaron.Context, params model.ModifyStudentStatusParams, sess session.Store) {
	if params.Status < 1 || params.Status > 2 || params.Id == 0 {
		c.JSON(200, RespError{false, "参数错误"})
		return
	}
	// 获取教练Id
	id := getCoachId(c, sess)
	if id == 0 {
		return
	}
	// 查找学员
	student := model.Student{Id: params.Id}
	ok, err := engine.Get(&student)
	if err != nil {
		LLog("engine.Get(&student) in ModifyStudentStatus()", err)
		Errors(c)
		return
	}
	if !ok {
		c.JSON(200, RespError{false, "不存在此ID"})
		return
	}
	if id != student.CoachId {
		c.JSON(200, RespError{false, "不能修改非自己学员的信息"})
		return
	}
	student.Sstatus = params.Status
	kvs := make(map[string]interface{})
	kvs["student"] = student
	commitWithDB(c, kvs, func(c *macaron.Context, kvs map[string]interface{}, s *xorm.Session) interface{} {
		v, _ := kvs["student"]
		student := v.(model.Student)
		// commitWithDB(c, student, func(c *macaron.Context, v interface{}, s *xorm.Session) interface{} {
		// 	student := v.(model.Student)
		num, err := s.Id(student.Id).Cols("Sstatus").Update(&student)
		if err != nil {
			LLog("s.Id(student.Id).Cols(Sstatus).Update(&student) in ModifyStudentStatus()", err)
			Errors(c)
			return nil
		}
		if num == 0 {
			c.JSON(200, RespError{false, "未做任何修改"})
			return nil
		}
		return RespStudent{true, student}
	})
}

// 处理学员绑定
func Binding(c *macaron.Context, param model.BindingParams, sess session.Store) {
	// 1-搜索记录，2-状态检查，3-更新数据
	if param.StudentId <= 0 {
		resultError(c, 7, "参数错误")
	}
	coachId := getCoachId(c, sess)
	if coachId == 0 {
		resultError(c, 6, "")
		return
	}
	student := model.Student{Id: param.StudentId, CoachId: coachId}
	has, err := engine.Get(&student)
	if err != nil {
		resultError(c, 6, "")
		return
	}
	if !has {
		resultError(c, 7, "没有此学员")
		return
	}
	if student.Bindstatus != 1 {
		resultError(c, 7, "此学员没有提交绑定申请")
		return
	}
	kvs := make(map[string]interface{})
	kvs["student"] = student
	commitWithDB(c, kvs, func(c *macaron.Context, kvs map[string]interface{}, s *xorm.Session) interface{} {
		v, _ := kvs["student"]
		student := v.(model.Student)
		student.Bindstatus = 2
		num, err := s.Id(student.Id).Cols("Bindstatus").Update(&student)
		if err != nil {
			LLog("s.Id(student.Id).Cols(Bindstatus).Update(&student) in Binding()", err)
			resultError(c, 6, "")
			return nil
		}
		if num == 0 {
			resultError(c, 7, "未做任何修改")
			return nil
		}
		return model.Result{Code: 1}
	})
}
