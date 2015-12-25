// 教练排课相关
package router

import (
	"errors"
	// "github.com/go-macaron/binding"
	"github.com/go-macaron/session"
	"github.com/go-xorm/xorm"
	"github.com/jinzhu/now"
	"github.com/leekchan/timeutil"
	"gopkg.in/macaron.v1"
	"log"
	"strconv"
	"time"
	"yj-server-golang/models"
)

// 排课
func CoachSchedule(c *macaron.Context, sess session.Store, params model.ScheduleParams) {
	// 参数检查
	if params.Day < 0 || params.Day > 6 {
		resultError(c, 3, "参数'day'错误")
		return
	}
	if params.Hour < 7 || params.Hour > 18 {
		resultError(c, 3, "参数'hour'错误")
		return
	}
	if params.Subject < 2 || params.Subject > 3 {
		resultError(c, 3, "参数'subject'错误")
		return
	}
	// 1 检查时间是否在课排程范围内
	// 2 检查是否可排程
	t := time.Now()
	d := timeutil.Timedelta{Days: time.Duration(params.Day)}
	td := t.Add(d.Duration())
	dateTime, err := time.ParseInLocation("2006-01-02 15:04:05", td.Format("2006-01-02 ")+strconv.Itoa(params.Hour)+":00:00", time.Local)
	// log.Println("1-", dateTime, dateTime.Unix())
	if err != nil {
		LLog("dateTime, err := time.Parse(2006-01-02 15:04:05, td.Format(2006-01-02 )+strconv.Itoa(params.Hour)+:00:00) in CoachSchedule()", err)
		resultError(c, 6, "")
		return
	}
	// 1
	if params.Day == 0 {
		now, err := time.ParseInLocation("2006-01-02 15:04:05", t.Format("2006-01-02 15:04:05"), time.Local)
		if err != nil {
			log.Println("2-", now)
			LLog("now, err := time.Parse(2006-01-02 15:04:05, t.Format(2006-01-02 15:04:05)) in CoachSchedule()", err)
			resultError(c, 6, "")
			return
		}
		if !dateTime.After(now) {
			resultError(c, 3, "无效的排课请求")
			return
		}
		//  排课
	}
	coachId := getCoachId(c, sess)
	if coachId == 0 {
		resultError(c, 6, "")
	}
	// tempSchedule := model.Schedule{CoachId: coachId, Datetime: dateTime.Unix()}
	tempSchedule := model.Schedule{CoachId: coachId, Datetime: model.JsonTime(dateTime)}
	// 2
	has, err := engine.Get(&tempSchedule)
	if err != nil {
		LLog("has, err := engine.Get(&tempSchedule) in CoachSchedule()", err)
		resultError(c, 6, "")
		return
	}
	if has && (tempSchedule.Status == 1 || tempSchedule.Status == 2) {
		resultError(c, 3, "已经添加过该时段")
		return
	}
	// schedule := model.Schedule{CoachId: coachId, Datetime: dateTime.Unix(), Subject: params.Subject, Status: 1}

	schedule := model.Schedule{CoachId: coachId, Datetime: model.JsonTime(dateTime), Subject: params.Subject, Status: 1}
	kvs := make(map[string]interface{})
	kvs["schedule"] = schedule
	commitWithDB(c, kvs, func(c *macaron.Context, kvs map[string]interface{}, s *xorm.Session) interface{} {
		v, _ := kvs["schedule"]
		schedule := v.(model.Schedule)
		s1 := model.Schedule{Id: 23}
		engine.Get(&s1)
		log.Printf("%#v", s1)
		num, err := s.Insert(&schedule)
		if err != nil {
			LLog("num, err := s.Insert(&schedule) in Schedule()", err)
			resultError(c, 6, "")
			return nil
		}
		if num == 0 {
			LLog("num, err := s.Insert(&schedule) in Schedule()", errors.New("num == 0"))
			resultError(c, 6, "")
			return nil
		}
		return model.Result{Code: 1, Data: schedule}
	})
}

// 修改排程
func ModifySchedule(c *macaron.Context, sess session.Store, params model.ModifyScheduleParams) {
	// 1-参数检查，2-排程是否存在，3-是否有人预约，4-是否在可修改的时间范围内

	if params.ScheduleId <= 0 {
		resultError(c, 3, "错误的课程ID")
		return
	}
	if params.Subject < 2 || params.Subject > 3 {
		resultError(c, 3, "参数'subject'错误")
		return
	}
	schedule := model.Schedule{Id: params.ScheduleId}
	has, err := engine.Get(&schedule)
	if err != nil {
		resultError(c, 6, "")
		return
	}
	if !has {
		resultError(c, 7, "课程不存在")
		return
	}
	dateTime, err := time.ParseInLocation("2006-01-02 15:04:05", time.Now().Format("2006-01-02 15:04:05"), time.Local)
	if err != nil {
		LLog("time.ParseInLocation(2006-01-02 15:04:05, td.Format(2006-01-02 15:04:05), time.Local) in CancelSchedule()", err)
		resultError(c, 6, "")
		return
	}
	log.Println(dateTime)
	log.Println(time.Time(schedule.Datetime))
	if !dateTime.Before(time.Time(schedule.Datetime)) {
		resultError(c, 7, "不能在课程开始后修改")
		return
	}
	if schedule.Status == 2 {
		resultError(c, 7, "此课程已经有人预约不可以修改")
		return
	}
	if schedule.Status == 3 {
		resultError(c, 7, "此课程已经被取消过不可以修改")
		return
	}
	schedule.Subject = params.Subject
	kvs := make(map[string]interface{})
	kvs["schedule"] = schedule
	commitWithDB(c, kvs, func(c *macaron.Context, kvs map[string]interface{}, s *xorm.Session) interface{} {
		v, _ := kvs["schedule"]
		schedule := v.(model.Schedule)
		num, err := s.Id(schedule.Id).Cols("Subject").Update(&schedule)
		if err != nil {
			LLog("num, err := s.Id(student.Id).Cols(Status).Update(&schedule)", err)
			resultError(c, 6, "")
			return nil
		}
		if num == 0 {
			resultError(c, 6, "")
			return nil
		}
		return model.Result{Code: 1, Data: schedule}
	})

}

// 取消排程
func CancelSchedule(c *macaron.Context, sess session.Store) {
	// 1-参数检查，2-排程是否存在，3-是否有人预约，4-是否在可取消的时间范围内
	scheduleId := c.ParamsInt64("id")
	if scheduleId <= 0 {
		resultError(c, 3, "错误的课程ID")
		return
	}
	schedule := model.Schedule{Id: scheduleId}
	has, err := engine.Get(&schedule)
	if err != nil {
		resultError(c, 6, "")
		return
	}
	if !has {
		resultError(c, 7, "课程不存在")
		return
	}

	dateTime, err := time.ParseInLocation("2006-01-02 15:04:05", time.Now().Format("2006-01-02 15:04:05"), time.Local)
	if err != nil {
		LLog("time.ParseInLocation(2006-01-02 15:04:05, td.Format(2006-01-02 15:04:05), time.Local) in CancelSchedule()", err)
		resultError(c, 6, "")
		return
	}
	if !dateTime.Before(time.Time(schedule.Datetime)) {
		resultError(c, 7, "不能在课程开始后修改")
		return
	}
	if schedule.Status == 2 {
		resultError(c, 7, "此课程已经有人预约不可以取消")
		return
	}
	if schedule.Status == 3 {
		resultError(c, 7, "此课程已经被取消不可以再次取消")
		return
	}
	schedule.Status = 3
	kvs := make(map[string]interface{})
	kvs["schedule"] = schedule
	commitWithDB(c, kvs, func(c *macaron.Context, kvs map[string]interface{}, s *xorm.Session) interface{} {
		v, _ := kvs["schedule"]
		schedule := v.(model.Schedule)
		num, err := s.Id(schedule.Id).Cols("Status").Update(&schedule)
		if err != nil {
			LLog("num, err := s.Id(student.Id).Cols(Status).Update(&schedule)", err)
			resultError(c, 6, "")
			return nil
		}
		if num == 0 {
			resultError(c, 6, "")
			return nil
		}
		return model.Result{Code: 1, Data: schedule}
	})
}

// 获取某一天的排程信息 ,参数day(0-6)
func GetSchedules(c *macaron.Context, sess session.Store) {
	// 1-参数检查，2-按日期获取排程
	day := c.QueryInt("day")
	if day < -30 || day > 6 {
		resultError(c, 3, "参数DAY错误")
		return
	}
	t := time.Now()
	if day != 0 {
		d := timeutil.Timedelta{Days: time.Duration(day)}
		t = t.Add(d.Duration())
	}
	coachId := getCoachId(c, sess)
	if coachId == 0 {
		resultError(c, 6, "")
		return
	}
	// start := now.New(t).BeginningOfDay().Unix()
	// end := now.New(t).EndOfDay().Unix()
	start := now.New(t).BeginningOfDay()
	end := now.New(t).EndOfDay()
	schedule := model.Schedule{CoachId: coachId}
	schedules := make([]model.Schedule, 0)
	err := engine.Where("datetime between ? and ? and status != ?", start, end, 3).Find(&schedules, schedule)
	if err != nil {
		LLog("err := engine.Where(datetime between ? and ?, start, end).Find(&schedules)", err)
		resultError(c, 6, "")
		return
	}
	result(c, schedules)
}

// 自动排课
func AutoScheule() {
	// 搜索教练表
	// 教练的排课表
	// 添加排课记录
	// 教练总数
	log.Println("开始自动排课")
	coach := model.Coach{}
	total, err := engine.Count(coach)
	if err != nil {
		LLog("engine.Where(id >?, 0).Count(coach) in AutoScheule()", err)
		return
	}
	var ii int64
	log.Println("教练数", total)
	ii = 0
	for i := 0; ii < total; {
		log.Println("i=", i)
		coachs := make([]model.Coach, 0)
		err := engine.Limit(50, i*50).Find(&coachs)
		if err != nil {
			LLog("err := engine.Limit(i*50, 50).Find(&coachs) in AutoScheule()", err)
			continue
		}
		for _, coach := range coachs {
			log.Println("coach.ID=", coach.Id)
			schedule := model.Schedule{CoachId: coach.Id}
			start := now.BeginningOfDay()
			end := now.EndOfDay()
			schedules := make([]model.Schedule, 0)
			err := engine.Where("datetime between ? and ? and status != ?", start, end, 3).Find(&schedules, schedule)
			if err != nil {
				LLog("err := engine.Where(datetime between ? and ?, start, end).Find(&schedules) in AutoScheule()", err)
				return
			}

			for _, schedule := range schedules {
				t := time.Time(schedule.Datetime)
				d := timeutil.Timedelta{Days: time.Duration(7)}
				date := model.JsonTime(t.Add(d.Duration()))
				schedule.Datetime = date
				// 开启事务
				session := engine.NewSession()
				err := session.Begin()
				defer session.Close()
				if err != nil {
					LLog("session.Begin() in AutoScheule()", err)
					continue
				}
				schedule.Id = 0
				schedule.Number = 0
				schedule.Status = 1
				schedule.Source = 1
				num, err := session.Insert(schedule)
				if err != nil {
					LLog("ssession.Insert(schedule) in AutoScheule()", err)
					session.Rollback()
					continue
				}
				if num == 0 {
					log.Println("num, err := session.Insert(schedule) num == 0 in AutoScheule()")
					session.Rollback()
					continue
				}
				err = session.Commit()
				if err != nil {
					LLog("session.Commit() in AutoScheule() in AutoScheule()", err)
					session.Rollback()
					continue
				}
			}
		}
		i = i + 50
		ii = int64(i)
	}
}
