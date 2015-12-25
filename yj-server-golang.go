package main

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/go-macaron/binding"
	"github.com/go-macaron/session"
	_ "github.com/go-macaron/session/redis"
	"gopkg.in/macaron.v1"
	"io/ioutil"
	"log"
	"net/http"
	"yj-server-golang/models"
	"yj-server-golang/router"
	"yj-server-golang/settings"
)

func main() {
	m := macaron.Classic()
	macaron.Env = macaron.PROD
	m.Use(macaron.Renderer())
	m.Post("/goBasic/coach/register/code", binding.Bind(model.RegisterParams{}), router.GetCode)
	m.Post("/goBasic/coach/register/register", binding.Bind(model.RegisterParams{}), router.RecordLoginDate, router.Register)
	m.Post("/goBasic/coach/register/login", binding.Bind(model.LoginParams{}), router.RecordLoginDate, router.Login)
	m.Post("/goBasic/coach/register/resetPassword", binding.Bind(model.RegisterParams{}), router.RecordLoginDate, router.ResetPassword)
	m.Group("/goBasic/coach/operation", func() {
		m.Put("/newSession", router.UpdateSession)                                                                 //更新 session
		m.Put("/modifyInfo", binding.Bind(model.ModifyCoachInfo{}), router.ModifyInfo)                             // 修改教练信息.
		m.Post("/addStudents", binding.Bind(model.AddStudentsParams{}), router.AddStudents)                        // 添加学员.
		m.Get("/getToken", router.GetToken)                                                                        // 获取Token
		m.Post("/pictureCallback", binding.Bind(model.PictureCallBackParams{}), router.PictureCallback)            // 上传图片回调.
		m.Post("/newClass", binding.Bind(model.Class{}), router.NewClass)                                          // 创建班型.
		m.Get("/getStudents", binding.Bind(model.GetStudentsParams{}), router.GetStudents)                         // 获取不同状态学员列表.
		m.Get("/getStudentInfo", router.GetStudentInfo)                                                            // 获取 学员信息.
		m.Put("/modifyRemrak", binding.Bind(model.ModifyRemrakParams{}), router.ModifyRemrak)                      // 修改学员备注.
		m.Get("/searchStudent", binding.Bind(model.SearchStudentParams{}), router.SearchStudent)                   // 搜索学员
		m.Put("/modifyStudentStatus", binding.Bind(model.ModifyStudentStatusParams{}), router.ModifyStudentStatus) // 修改学员状态.

		m.Delete("/deleteStudent/:id", router.DeleteStudent) // 删除学员

		m.Put("/binding", binding.Bind(model.BindingParams{}), router.Binding) // 同意学员绑定

		m.Get("/getPictures", router.GetPictures)                                                   // 获取图片
		m.Get("/reload", router.Reload)                                                             // 加载个人信息
		m.Delete("/deletePicture/:id", router.DeletePicture)                                        // 删除图片
		m.Get("/getAllStudents", router.GetAllStudents)                                             //获取所有学员
		m.Put("/modifyClass", binding.Bind(model.Class{}), router.ModifyCalss)                      //修改班型.
		m.Get("/searchDriveSchool", router.SearchDriveSchool)                                       //搜索驾校
		m.Post("/applyToCoach", router.ApplyToCoach)                                                //提交审核
		m.Get("/getClassInfo", router.GetClassInfo)                                                 //搜索驾校
		m.Post("/schedule", binding.Bind(model.ScheduleParams{}), router.CoachSchedule)             // 排课
		m.Put("/modifySchedule", binding.Bind(model.ModifyScheduleParams{}), router.ModifySchedule) // 修改排程
		m.Delete("/cancelSchedule/:id", router.CancelSchedule)                                      // 取消排程
		m.Get("/getSchedules", router.GetSchedules)                                                 // 获取排程

	}, router.Authorization, router.Validity, router.RecordLoginDate)

	m.Use(session.Sessioner(session.Options{
		Provider: "redis",
		// e.g.: network=tcp,addr=127.0.0.1:6379,password=macaron,db=0,pool_size=100,idle_timeout=180,prefix=session:
		// ProviderConfig: "addr=" + config.RedisAdrr + ",password=footbakk-YJKJ,db=sess,prefix=session:",
		ProviderConfig: "addr=" + config.RedisAdrr + ",password=testredis,db=sess,prefix=session:",
		CookieName:     "connect.sid",
		Gclifetime:     3600 * 24 * 40,
		Maxlifetime:    3600 * 24 * 40,
		CookieLifeTime: 3600 * 24 * 40,
		IDLength:       24,
		CookiePath:     "/",
	}))
	// m.Run()
	pool := x509.NewCertPool()
	caCertPath := "/home/blackcat/ca.crt"
	// caCertPath := "ca.crt"

	caCrt, err := ioutil.ReadFile(caCertPath)
	if err != nil {
		log.Println("ReadFile err:", err)
		return
	}
	pool.AppendCertsFromPEM(caCrt)

	s := &http.Server{
		Addr:    ":4000",
		Handler: m,
		TLSConfig: &tls.Config{
			ClientCAs:  pool,
			ClientAuth: tls.RequireAndVerifyClientCert,
		},
	}
	err = s.ListenAndServeTLS("/home/blackcat/server.crt", "/home/blackcat/server.key")
	// err = s.ListenAndServeTLS("server.crt", "server.key")
	if err != nil {
		log.Println("ListenAndServeTLS err:", err)
	}
}
