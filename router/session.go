package router

import (
	"github.com/go-macaron/session"
	"gopkg.in/macaron.v1"
	// "log"
)

// 更新session
func UpdateSession(c *macaron.Context, sess session.Store) {
	// coach := sess.Get("coach")
	// recentLoginDate := sess.Get("recentLoginDate")
	// id := sess.ID()
	// log.Println(id)
	ccopy := c
	sess.RegenerateId(c)
	err := sess.Destory(ccopy)
	if err != nil {
		LLog("sess.Destory(ccopy) in UpdateSession", err)
		Errors(c)
		return
	}
	c.JSON(200, RespSuccese{true})
}
