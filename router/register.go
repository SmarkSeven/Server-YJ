package router

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	// "github.com/elgs/cron"
	"github.com/go-macaron/session"
	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/core"
	"github.com/go-xorm/xorm"
	"github.com/leekchan/timeutil"
	"gopkg.in/macaron.v1"
	"qiniupkg.com/api.v7/kodo"
	"regexp"
	"strings"
	"time"
	"yj-server-golang/models"
	// . "yj-server-golang/models"
	"yj-server-golang/settings"
	"yj-server-golang/tool"
)

// 给客户端放回成功的信息
type RespSuccese struct {
	Success bool `json:"success"`
}

type RespInfo struct {
	Success  bool        `json:"success"`
	UserInfo model.Coach `json:"userInfo"`
}

// 给客户端返回失败的信息
type RespError struct {
	Success   bool   `json:"success"`
	ErrorInfo string `json:"errorInfo"`
}

type Cook struct {
	OriginalMaxage interface{} `json:"originalMaxAge"`
	Expires        interface{} `json:"expires"`
	HttpOnly       bool        `json:"httpOnly"`
	Path           string      `json:"path"`
}

// 全局变量
var engine *xorm.Engine

// 获取验证码
func GetCode(c *macaron.Context, params model.RegisterParams, sess session.Store) {
	if sess == nil {
		return
	}
	var phone = params.Phone
	if !CheckPhone(phone) {
		c.JSON(200, RespError{false, "手机号格式不正确"})
		return
	}
	code := tool.Rand(1000, 10000)
	datas := []string{code, "5"}
	result, err := Sms(phone, "50306", datas)
	if err != nil {
		Errors(c)
		return
	}
	LLog(result, nil)
	resultSMS := model.ResultSMS{}
	err = json.Unmarshal([]byte(result), &resultSMS)
	if err != nil {
		Errors(c)
		return
	}
	if !strings.EqualFold(resultSMS.StatusCode, "000000") {
		c.JSON(200, RespError{false, "发送验证码失败"})
		return
	}

	//将验证码存入session中,验证码有效期为5分钟
	verifyCodeTemp := model.VerifyCodeTemp{Phone: phone, Code: code}
	t := time.Now()
	td := timeutil.Timedelta{Days: 0, Minutes: time.Duration(5), Seconds: 0}
	t = t.Add(td.Duration())
	verifyCodeTemp.DeadLine = t.Format("2006-01-02 15:04:05")
	sess.Set("verifyCodeTemp", verifyCodeTemp)
	c.JSON(200, RespSuccese{true})
}

// 注册用户
func Register(c *macaron.Context, params model.RegisterParams, sess session.Store) {
	sid := c.GetCookie("connect.sid")
	if len(sid) <= 0 { // 没有cookie
		c.JSON(200, RespError{false, "没有有效的凭证"})
		return
	}
	// 请求参数
	var phone = params.Phone
	var code = params.Code

	if !CheckPhone(phone) {
		c.JSON(200, RespError{false, "手机号格式不正确"})
		return
	}
	if len(params.Password) < 6 {
		c.JSON(200, RespError{false, "密码长度不够"})
		return
	}
	// 检查手机号是否被注册
	user := model.User{Phone: phone}
	ok, err := engine.Get(&user)
	if err != nil {
		c.JSON(200, RespError{false, "系统错误"})
		return
	}
	if ok {
		switch user.Role {
		case 1:
			c.JSON(200, RespError{false, "手机号已被注册为教练"})
			return
		case 2:
			c.JSON(200, RespError{false, "手机号已被注册为学员"})
			return
		}
		c.JSON(200, RespError{false, "手机号已经被注册"})
		return
	}
	kvs := make(map[string]interface{})
	if sess.Get("verifyCodeTemp") == nil {
		c.JSON(200, RespError{false, "无效的凭证"})
		sess.Destory(c)
		return
	}
	kvs = map[string]interface{}(sess.Get("verifyCodeTemp").(map[string]interface{}))
	if kvs == nil {
		c.JSON(200, RespError{false, "无效的凭证"})
		sess.Destory(c)
		return
	}
	verifyCodeTemp := model.VerifyCodeTemp{Code: kvs["code"].(string), Phone: kvs["phone"].(string), DeadLine: kvs["deadLine"].(string)}
	if !strings.EqualFold(phone, verifyCodeTemp.Phone) {
		c.JSON(200, RespError{false, "手机号不匹配"})
		return
	}
	if !strings.EqualFold(code, verifyCodeTemp.Code) {
		c.JSON(200, RespError{false, "验证码错误"})
		return
	}
	// 验证码是否过期
	t, _ := time.Parse("2006-01-02 15:04:05", time.Now().Format("2006-01-02 15:04:05")) // 请求发起时间
	deadTime, _ := time.Parse("2006-01-02 15:04:05", verifyCodeTemp.DeadLine)
	if t.After(deadTime) {
		c.JSON(200, RespError{false, "验证码已失效"})
		return
	}

	origData, err := base64.StdEncoding.DecodeString(params.Password)
	if err != nil {
		LLog("base64.StdEncoding.DecodeString(params.Password) in Register()", err)
		Errors(c)
		return
	}
	ciphertext, err := tool.RsaDecrypt(origData)
	if err != nil {
		LLog("tool.RsaDecrypt(origData) in Register()", err)
		c.JSON(200, RespError{false, "错误的密码"})
		return
	}
	h := md5.New()
	h.Write(ciphertext)
	// coach.Password = hex.EncodeToString(h.Sum(nil))
	coach := model.Coach{Phone: phone, Password: hex.EncodeToString(h.Sum(nil))}
	// 开启事务
	session := engine.NewSession()
	err = session.Begin()
	defer session.Close()
	if err != nil {
		LLog("session.Begin()", err)
		Errors(c)
		return
	}
	// 添加coach
	num, err := session.Insert(&coach)
	if err != nil {
		LLog("num, err := session.Insert(&coach)", err)
		c.JSON(200, RespError{false, "系统错误"})
		session.Rollback()
		return
	}
	if num == 0 {
		c.JSON(200, RespError{false, "系统错误"})
		session.Rollback()
		return
	}
	// user信息
	user.RoleId = coach.Id
	user.Name = coach.Name
	user.Phone = coach.Phone
	user.Password = coach.Password
	user.Role = 1

	// 添加user
	num, err = session.Insert(&user)
	if err != nil {
		LLog("session.Insert(&user)", err)
		Errors(c)
		session.Rollback()
		return
	}

	if num == 0 {
		LLog("session.Insert(&user)", err)
		Errors(c)
		session.Rollback()
		return
	}
	coach.Userid = user.Id
	num, err = session.Id(coach.Id).Update(&coach)
	if err != nil {
		LLog("session.Id(coach.Id).Update(&coach)", err)
		Errors(c)
		session.Rollback()
		return
	}
	if num == 0 {
		Errors(c)
		session.Rollback()
		return
	}
	err = session.Commit()
	if err != nil {
		LLog("session.Commit()", err)
		c.JSON(200, RespError{false, "添加教练失败"})
		session.Rollback()
		return
	}
	err = sess.Delete("verifyCodeTemp")
	if err != nil {
		Errors(c)
		return
	}
	sess.Set("coach", coach)
	if err != nil {
		Errors(c)
		return
	}
	coach.Password = ""
	c.JSON(200, RespInfo{true, coach})
}

// 用户登录
func Login(c *macaron.Context, params model.LoginParams, sess session.Store) {
	var phone = params.Phone
	if !CheckPhone(phone) {
		c.JSON(200, RespError{false, "手机号格式不正确"})
		return
	}
	if len(params.Password) == 0 {
		c.JSON(200, RespError{false, "密码为空"})
		return
	}
	origData, err := base64.StdEncoding.DecodeString(params.Password)
	if err != nil {
		LLog("base64.StdEncoding.DecodeString(params.Password) in Login()", err)
		Errors(c)
		return
	}
	ciphertext, err := tool.RsaDecrypt(origData)
	if err != nil {
		LLog("tool.RsaDecrypt(origData) in Login()", err)
		c.JSON(200, RespError{false, "错误的密码"})
		return
	}
	h := md5.New()
	h.Write(ciphertext)
	// h := md5.New()
	// h.Write([]byte(params.Password))
	coach := model.Coach{Phone: phone, Password: hex.EncodeToString(h.Sum(nil))}
	// 核对数据库
	ok, err := engine.Get(&coach)
	if err != nil {
		LLog("engine.Get(&coach)", err)
		Errors(c)
		return
	}
	if !ok {
		coach2 := model.Coach{Phone: phone}
		ok, err = engine.Get(&coach2)
		if err != nil {
			LLog("engine.Get(&coach)", err)
			Errors(c)
			return
		}
		if !ok {
			c.JSON(200, RespError{false, "手机号没注册"})
			return
		}
		c.JSON(200, RespError{false, "密码错误"})
		return
	}
	sid := c.GetCookie("connect.sid")
	// 带有cookie
	if len(sid) > 0 {
		// 检查请求参数和cookie是否匹配
		kvs := make(map[string]interface{})
		if sess.Get("coach") == nil {
			sess.Set("coach", coach)
			c.JSON(200, RespInfo{true, coach})
			return
		}
		kvs = map[string]interface{}(sess.Get("coach").(map[string]interface{}))
		if kvs == nil {
			c.JSON(200, RespError{false, "无效的凭证"})
			sess.Destory(c)
			return
		}
		if !strings.EqualFold(phone, kvs["phone"].(string)) {
			c.JSON(200, RespError{false, "无效的凭证"})
			return
		}
		// 更新session内容
		sess.Set("coach", coach)
		c.JSON(200, RespInfo{true, coach})
		return
	}
	// 没有带cookie
	// 更新session内容
	sess.Set("coach", coach)
	coach.Password = ""
	c.JSON(200, RespInfo{true, coach})
	return
}

//修改密码
func ResetPassword(c *macaron.Context, params model.RegisterParams, sess session.Store) {
	var phone = params.Phone
	var code = params.Code
	if sess == nil {
		resultError(c, 6, "")
		return
	}
	coach := model.Coach{Phone: phone}

	if !CheckPhone(phone) {
		c.JSON(200, RespError{false, "手机格式不正确"})
		return
	}
	ok, err := engine.Get(&coach)
	if err != nil {
		LLog("engine.Get(&coach) in ResetPassword()", err)
		resultError(c, 6, "")
		return
	}
	if !ok {
		resultError(c, 7, "手机号没注册")
		return
	}
	origData, err := base64.StdEncoding.DecodeString(params.Password)
	if err != nil {
		LLog("base64.StdEncoding.DecodeString(params.Password) in ResetPassword()", err)
		resultError(c, 6, "")
		return
	}
	ciphertext, err := tool.RsaDecrypt(origData)
	if err != nil {
		LLog("tool.RsaDecrypt(origData) in ResetPassword()", err)
		resultError(c, 6, "")
		return
	}
	h := md5.New()
	h.Write(ciphertext)
	coach.Password = hex.EncodeToString(h.Sum(nil))
	kvs := make(map[string]interface{})
	if sess.Get("verifyCodeTemp") == nil {
		resultError(c, 7, "请先获取验证码")
		return
	}
	kvs = map[string]interface{}(sess.Get("verifyCodeTemp").(map[string]interface{}))
	if kvs == nil {
		resultError(c, 7, "请先获取验证码")
		return
	}
	verifyCodeTemp := model.VerifyCodeTemp{Code: kvs["code"].(string), Phone: kvs["phone"].(string), DeadLine: kvs["deadLine"].(string)}
	if !strings.EqualFold(phone, verifyCodeTemp.Phone) {
		resultError(c, 3, "手机号不匹配")
		return
	}
	if !strings.EqualFold(code, verifyCodeTemp.Code) {
		resultError(c, 3, "验证码错误")
		return
	}
	t, _ := time.Parse("2006-01-02 15:04:05", time.Now().Format("2006-01-02 15:04:05")) // 请求发起时间
	deadTime, _ := time.Parse("2006-01-02 15:04:05", verifyCodeTemp.DeadLine)
	if t.After(deadTime) {
		resultError(c, 7, "验证码已失效")
		return
	}
	session := engine.NewSession()
	err = session.Begin()
	defer session.Close()
	if err != nil {
		LLog("session.Begin() in ResetPassWord()", err)
		resultError(c, 6, "")
		return
	}
	// 更新密码
	num, err := session.Id(coach.Id).Update(&coach)
	if err != nil {
		LLog("Update(&coach) in ResetPassWord()", err)
		resultError(c, 6, "")
		session.Rollback()
		return
	}
	if num == 0 {
		resultError(c, 7, "为做任何修改")
		session.Rollback()
		return
	}
	user := model.User{RoleId: coach.Id}
	ok, err = engine.Get(&user)
	if err != nil {
		LLog("engine.Get(&user) in ResetPassWord()", err)
		resultError(c, 6, "")
		return
	}
	if !ok {
		resultError(c, 7, "手机号没注册")
		return
	}
	user.Password = coach.Password

	num, err = session.Id(user.Id).Update(&user)
	if err != nil {
		LLog("Update(&user) in ResetPassWord()", err)
		resultError(c, 6, "")
		session.Rollback()
		return
	}
	if num == 0 {
		resultError(c, 7, "为做任何修改")
		session.Rollback()
		return
	}

	err = session.Commit()
	if err != nil {
		LLog("session.Commit() in ResetPassWord", err)
		resultError(c, 7, "重置密码失败")
		session.Rollback()
		return
	}
	err = sess.Delete("verifyCodeTemp")
	if err != nil {
		LLog("sess.Delete(verifyCodeTemp) in ResetPassWord", err)
		resultError(c, 6, "")
		return
	}
	if ok {
		sess.Set("coach", coach)
	}
	result(c, coach)

}

// 加载教练信息
func Reload(c *macaron.Context, sess session.Store) {
	id := getCoachId(c, sess)
	if id == 0 {
		return
	}
	coach := model.Coach{Id: id}
	ok, err := engine.Get(&coach)
	if err != nil {
		Errors(c)
		return
	}
	if !ok {
		c.JSON(200, RespError{false, "获取信息失败"})
		return
	}
	coach.Password = ""
	c.JSON(200, RespInfo{true, coach})
}

func CheckPhone(phone string) bool {
	if m, _ := regexp.MatchString(`^(1[3|7|4|5|8][0-9]\d{4,8})$`, phone); !m {
		return false
	}
	if len(phone) != 11 {
		return false
	}
	return true
}

func init() {
	var err error
	engine, err = xorm.NewEngine("mysql", "root:mysql@tcp("+config.MySQLIP+")/helpdrive?charset=utf8&loc=Asia%2FShanghai")
	if err != nil {
		LLog("xorm.NewEngine in regidter.go init()", err)
	}
	location, err := time.LoadLocation("Asia/Shanghai")
	engine.TZLocation = location
	engine.SetMaxConns(10)
	engine.SetMaxIdleConns(10)

	err = engine.Ping()
	if err != nil {
		LLog("err = engine.Ping() regidter.go init()", err)
	}
	// 设置映射规则
	engine.SetMapper(core.SameMapper{})
	kodo.SetMac(config.QACCESS_KEY, config.QSECRET_KEY)

	// 定时任务
	// c := cron.New()
	// c.AddFunc("0 10 17 * * * ", AutoScheule)
	// c.Start()
}
