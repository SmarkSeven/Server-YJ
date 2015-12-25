package config

import (
	"log"
	"os"
)

var Env string // 项目运行环境
// var MySQLIP = "120.24.180.250:3306" //mysql服务器IP地址
// var PushIP string = "120.25.227.58" //push服务器的IP地址
// var RedisAdrr string = "120.24.180.250:6379"

var MySQLIP = "192.168.31.202:3306"  //mysql服务器IP地址
var PushIP string = "192.168.31.202" //push服务器的IP地址  120.25.227.58
var RedisAdrr string = "192.168.31.202:6379"
var MySQLUserName string  //mysql用户名
var MySQLPassword string  //mysql密码
var PackageName string    //包名
var DateBaseName string   //数据库，默认链接到learndrive
var AppKey string = "lja" //
var AppSecret string = "lja123456"
var PushMethod string = "POST"                                                //请求推送消息方式
var PushPath string = "/push/send/app"                                        //推送请求路径
var PushMessageUrl string = "http://" + PushIP + ":80" + "/push/send/app"     // 推送App消息
var PushSmsVerifyUrl string = "http://" + PushIP + ":80" + "/push/sms/verify" // 推送短信

//七牛配置
var QACCESS_KEY = "x2KJmdepFhAJpgGOVMipefEH6n2dOz_akFo4wQ9N"
var QSECRET_KEY = "_hVm15zwG4DSQFXzGWHQbRBSrv7-35an9uzy-QC8"

// 图片前缀
var AvatorPrex = "http://7xjnv4.com2.z0.glb.qiniucdn.com/"
var IdentityPrex = "http://7xjiyi.com2.z0.glb.qiniucdn.com/"
var CoachLisencePrex = "http://7xjjmt.com2.z0.glb.qiniucdn.com/"
var DirveLisencePrex = "http://7xjjmp.com2.z0.glb.qiniucdn.com/"
var SitePicturePrex = "http://7xoe3j.com2.z0.glb.qiniucdn.com/"
var CarPicturPrex = "http://7xoe3h.com2.z0.glb.qiniucdn.com/"

//
var DeletePushMethod string = "DELETE"
var DeletePushPath string = "/push"

//添加预约，延迟推送时间
var LocalPushDelay int64 = 10
var DevelopmentPushDelay int64 = 5
var TestingPushDelay int64 = 5
var TestingZJPushDelay int64 = 5
var DevelopmentZJBPushDelay int64 = 5

var Domain = "7xjiyi.com2.z0.glb.qiniucdn.com"

// var ConfConst ConfigSetting //在init()中初始化
func init() {
	Env = os.Getenv("GO_ENV")
	if Env == "" {
		log.Printf("get GoEnv value failed")
		Env = "production"
	}
}
