package router

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"github.com/go-macaron/session"
	"github.com/go-xorm/xorm"
	"gopkg.in/macaron.v1"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
	. "yj-server-golang/models"
)

// 错误日志输出
func LLog(d string, err error) {
	log.Printf(d+"%#v", err)
}

// 返回错误
func Errors(c *macaron.Context) {
	c.JSON(200, RespError{false, "系统错误"})
}

// 返回错误  1-成功，2-错误的cookie，3-参数错误，4-无效的cookie，5-访问时间问题，6－系统错误，7-其它
func resultError(c *macaron.Context, code int, errorInfo string) {
	if len(errorInfo) == 0 {
		c.JSON(200, Result{Code: code})
	} else {
		c.JSON(200, Result{Code: code, ErrorInfo: errorInfo})
	}

}
func result(c *macaron.Context, data interface{}) {
	c.JSON(200, Result{Code: 1, Data: data})
}

// 返回错误远影
func ErrorJson(c *macaron.Context, errorInfo string) {
	c.JSON(200, RespError{false, errorInfo})
}

// 获取教练Id
func getCoachId(c *macaron.Context, sess session.Store) (id int64) {
	kvs := make(map[string]interface{})
	if sess.Get("coach") == nil {
		c.JSON(200, RespError{true, "没有登录"})
		return id
	}
	kvs = map[string]interface{}(sess.Get("coach").(map[string]interface{}))
	if kvs == nil {
		c.JSON(200, RespError{false, "没有登录"})
		return id
	}
	id = int64(kvs["id"].(float64))
	if id == 0 {
		Errors(c)
	}
	return id
}

// 获取coach的userid
func getuserId(c *macaron.Context, sess session.Store) (id int64) {
	kvs := make(map[string]interface{})
	if sess.Get("coach") == nil {
		c.JSON(200, RespError{true, "没有登录"})
		return id
	}
	kvs = map[string]interface{}(sess.Get("coach").(map[string]interface{}))
	if kvs == nil {
		c.JSON(200, RespError{false, "没有登录"})
		return id
	}
	id = int64(kvs["userid"].(float64))
	if id == 0 {
		Errors(c)
	}
	return id
}

// 数据库数据处理
func commitWithDB(c *macaron.Context, kvs map[string]interface{}, fn func(c *macaron.Context, kvs map[string]interface{}, s *xorm.Session) interface{}) {
	// 开启事务
	session := engine.NewSession()
	err := session.Begin()
	defer session.Close()
	if err != nil {
		LLog("session.Begin() in commitWithDB", err)
		Errors(c)
		return
	}
	//数据处理
	respData := fn(c, kvs, session)
	if respData == nil {
		session.Rollback()
		return
	}
	// 提交事务
	err = session.Commit()
	if err != nil {
		LLog("session.Commit() in commitWithDB", err)
		session.Rollback()
		Errors(c)
		return
	}
	c.JSON(200, respData)
}

// 数据库数据处理
func commitWithDB2(c *macaron.Context, kvs map[string]interface{}, fn func(c *macaron.Context, kvs map[string]interface{}, s *xorm.Session) bool, fn2 func(c *macaron.Context, kvs map[string]interface{}, s *xorm.Session) interface{}) {
	// 开启事务
	session := engine.NewSession()
	err := session.Begin()
	defer session.Close()
	if err != nil {
		LLog("session.Begin() in commitWithDB2", err)
		Errors(c)
		return
	}
	//数据处理
	ok := fn(c, kvs, session)
	if !ok {
		session.Rollback()
		return
	}
	// 提交事务
	err = session.Commit()
	if err != nil {
		LLog("session.Commit() in commitWithDB2", err)
		session.Rollback()
		Errors(c)
		return
	}
	respData := fn2(c, kvs, session)
	if respData == nil {
		return
	}
	c.JSON(200, respData)
}

// 发送短信
func Sms(to, templateId string, datas []string) (result string, err error) {
	type SmsCodeParams struct {
		To         string   `json:"to"`
		AppId      string   `json:"appId"`
		TemplateId string   `json:"templateId"`
		Datas      []string `json:"datas"`
	}

	client := &http.Client{Transport: nil}
	sid := "aaf98f894e3e5b81014e4885e6660a51"
	token := "6af59c5c9c624089ae34f248270211e9"
	timeNow := time.Now().Format("20060102150405")
	h := md5.New()
	h.Write([]byte(sid + token + timeNow))
	sig := hex.EncodeToString(h.Sum(nil))
	sig = strings.ToUpper(sig)

	uri := "https://app.cloopen.com:8883/2013-12-26/Accounts/" + sid + "/SMS/TemplateSMS?sig=" + sig

	smsCodeParams := SmsCodeParams{}
	smsCodeParams.To = to
	smsCodeParams.AppId = "8a48b5514fba2f87014fbf53fcb00d57"
	smsCodeParams.TemplateId = templateId
	smsCodeParams.Datas = datas

	b, err := json.Marshal(smsCodeParams)
	if err != nil {
		return result, err
	}

	v := string(b)
	req, err := http.NewRequest("POST", uri, strings.NewReader(v))
	if err != nil {
		return result, err
	}
	authorization := sid + ":" + timeNow
	authorization = base64.StdEncoding.EncodeToString([]byte(authorization))
	req.Header.Set("Content-Type", "application/json;charset=utf-8")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Length", strconv.Itoa(len(v)))
	req.Header.Set("Authorization", authorization)

	resp, err := client.Do(req)
	defer resp.Body.Close()
	if err != nil {
		return result, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return result, err
	}
	result = string(body)
	return result, nil
}
