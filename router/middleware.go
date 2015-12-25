// 通用中间件
package router

import (
	"github.com/go-macaron/session"
	"github.com/leekchan/timeutil"
	"gopkg.in/macaron.v1"
	"time"
)

// 授权
func Authorization(c *macaron.Context, sess session.Store) {
	// 1.检测cookie对应的session中有无coach信息
	if sess == nil {
		return
	}
	if sess.Get("coach") == nil {
		err := sess.Destory(c)
		if err != nil {
			LLog("", err)
			Errors(c)
			return
		}
		c.JSON(200, RespError{false, "没有权限."})
		return
	}

}

// 记录最后与服务器会话的时间
func RecordLoginDate(c *macaron.Context, sess session.Store) {
	if sess == nil {
		return
	}
	t := time.Now().Format("2006-01-02 15:05:04")
	err := sess.Set("recentLoginDate", t)
	if err != nil {
		LLog("ess.Set(recentLoginDate, t) in RecordDate()", err)
		Errors(c)
		return
	}
}

// 验证session是否在有效期内
func Validity(c *macaron.Context, sess session.Store) {
	if sess.Get("recentLoginDate") == nil {
		c.JSON(200, RespError{true, "请退出后重新登录"})
		return
	}
	date := string(sess.Get("recentLoginDate").(string))
	t, err := time.Parse("2006-01-02 15:04:05", date) //上次登录时间
	if err != nil {
		LLog("time.Parse(2006-01-02 15:04:05, date) in Validity()", err)
		Errors(c)
		return
	}
	td := timeutil.Timedelta{Days: time.Duration(40), Minutes: 0, Seconds: 0}
	today, _ := time.Parse("2006-01-02 15:04:05", time.Now().Format("2006-01-02 15:04:05")) // 请求发起时间
	deadline := t.Add(td.Duration())
	if !today.Before(deadline) {
		c.JSON(200, RespError{false, "登录已过有效期,请退出后重新登录"})
		return
	}
	err = sess.Set("recentLoginDate", today.Format("2006-01-02 15:04:05"))
	if err != nil {
		LLog("ess.Set(recentLoginDate, t) in Validity()", err)
		Errors(c)
		return
	}
}
