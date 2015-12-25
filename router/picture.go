package router

import (
	"errors"
	"github.com/go-macaron/session"
	"github.com/go-xorm/xorm"
	"golang.org/x/net/context"
	"gopkg.in/macaron.v1"
	"qiniupkg.com/api.v7/kodo"
	"strconv"
	"strings"
	"time"
	"yj-server-golang/models"
	"yj-server-golang/settings"
)

type RespToken struct {
	Success bool   `json:"success"`
	Key     string `json:"key"`
	Token   string `json:"token"`
}

type RespData struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
}

type UpImage struct {
	Scope     string `json:"scope"`
	Key       string `json:"key"`
	ImageType int    `json:"imageType"`
}

type RespPictures struct {
	Success  bool            `json:"success"`
	Pictures []model.Picture `json:"pictures"`
}

// type :1-头像 2-身份证，3-教练证，4-驾驶证，5-场地图片，6-车辆照片
func uptoken(imageType int, id int64) (policy kodo.PutPolicy, key string, err error) {
	var bucketName string
	// var key string
	switch imageType {
	case 1:
		bucketName = "avator"
		key = config.Env + "/" + strconv.FormatInt(id, 10)
	case 2:
		bucketName = "identity"
		key = config.Env + "/" + strconv.FormatInt(id, 10)
	case 3:
		bucketName = "coachlisence"
		key = config.Env + "/" + strconv.FormatInt(id, 10)
	case 4:
		bucketName = "drivelisence"
		key = config.Env + "/" + strconv.FormatInt(id, 10)
	case 5:
		bucketName = "site"
		pictures := make([]model.Picture, 0)
		picture := model.Picture{UserId: id, ImageType: 5}
		err = engine.Find(&pictures, &picture)
		if err != nil {
			LLog("case 5: engine.Find(&pictures, &picture) in uptoken()", err)
			return policy, key, errors.New("系统错误")
		}
		if len(pictures) >= 4 {
			return policy, key, errors.New("最多只能上传4张图片")
		}
		t := time.Now()
		//图片地址
		key = config.Env + "/" + strconv.FormatInt(id, 10) + "_" + t.Format("2006_01_02_15_05_04_") + strconv.Itoa(t.Nanosecond())
	case 6:
		bucketName = "carpicture"
		pictures := make([]model.Picture, 0)
		picture := model.Picture{UserId: id, ImageType: 6}
		err := engine.Find(&pictures, &picture)

		if err != nil {
			LLog("case 6: engine.Find(&pictures, &picture) in uptoken()", err)
			return policy, key, errors.New("系统错误")
		}
		if len(pictures) >= 4 {
			return policy, key, errors.New("最多只能上传4张图片")
		}
		t := time.Now()
		//时间戳到具体显示的转化
		key = config.Env + "/" + strconv.FormatInt(id, 10) + "_" + t.Format("2006_01_02_15_05_04_") + strconv.Itoa(t.Nanosecond())
	}
	policy = kodo.PutPolicy{
		Scope: bucketName + ":" + key,
	}
	return policy, key, nil

}

// 获取Token
func GetToken(c *macaron.Context, sess session.Store) {
	imageType := c.QueryInt("imageType")
	if imageType >= 2 && imageType <= 4 {
		coachId := getCoachId(c, sess)
		coach := model.Coach{Id: coachId}
		_, err := engine.Get(&coach)
		if err != nil {
			Errors(c)
			return
		}
		if coach.Status == 2 {
			c.JSON(200, RespError{false, "您已经通过认证"})
			return
		}
		if coach.Status == 3 {
			c.JSON(200, RespError{false, "您的申请正在审核中"})
			return
		}
		if coach.Status == 4 {
			c.JSON(200, RespError{false, "您的申请未通过审核，原因: " + coach.Remark})
			return
		}

	}
	if imageType < 1 || imageType > 6 {
		c.JSON(200, RespError{false, "参数错误"})
		return
	}
	// 检查是否有未处理的图片
	if sess.Get("upImage") != nil {
		c.JSON(200, RespError{false, "有未处理的图片"})
		return
	}
	client := getClient()
	id := getuserId(c, sess)
	if id == 0 {
		return
	}
	policy, key, err := uptoken(imageType, id)
	if err != nil {
		c.JSON(200, RespError{false, err.Error()})
		return
	}
	upImage := UpImage{Scope: policy.Scope, Key: key, ImageType: imageType}
	err = sess.Set("upImage", upImage)
	if err != nil {
		Errors(c)
		return
	}
	c.JSON(200, RespToken{true, key, client.MakeUptoken(&policy)})
}

// 图片上传完成后的回调  99-上传失败
func PictureCallback(c *macaron.Context, param model.PictureCallBackParams, sess session.Store) {
	imageType := param.ImageType
	if imageType == 99 {
		err := sess.Delete("upImage")
		if err != nil {
			LLog("sess.Delete(upImage) in PictureCallback()", err)
			// Errors(c)
			c.JSON(200, 1)
			return
		}
		c.JSON(200, RespSuccese{true})
		return
	}
	// 检查参数
	if imageType < 1 || imageType > 6 {
		c.JSON(200, RespError{false, "没有此类型的图片"})
		return
	}
	// 获取session中的信息
	kvs := make(map[string]interface{})
	if sess.Get("upImage") == nil {
		c.JSON(200, RespError{false, "没有上传过图片"})
		return
	}
	kvs = map[string]interface{}(sess.Get("upImage").(map[string]interface{}))
	if kvs == nil {
		c.JSON(200, RespError{false, "没有上传过图片"})
		return
	}
	upImage := UpImage{Scope: kvs["scope"].(string), Key: kvs["key"].(string), ImageType: int(kvs["imageType"].(float64))}
	if imageType != upImage.ImageType {
		c.JSON(200, RespError{false, "没有上传过此类型的图片"})
		return
	}
	// 获取用户ID
	coachId := getCoachId(c, sess)
	if coachId == 0 {
		return
	}
	id := getuserId(c, sess)
	if id == 0 {
		return
	}
	// 开启事务
	session := engine.NewSession()
	err := session.Begin()
	defer session.Close()
	if err != nil {
		LLog("session.Begin() in PictureCallback()", err)
		// Errors(c)
		c.JSON(200, 2)
		return
	}
	// 添加时间
	var picture model.Picture
	var data interface{}
	// var num int64 = -1
	switch imageType {
	case 1:
		url := config.AvatorPrex + upImage.Key
		if !HasPicture("avator", upImage.Key) {
			c.JSON(200, RespError{false, "图片不存在"})
			return
		}
		data = url
		coach := model.Coach{Id: coachId, Avator: url}
		_, err = session.Id(coachId).Cols("Avator").Update(&coach)
	case 2:
		url := config.IdentityPrex + upImage.Key
		if !HasPicture("identity", upImage.Key) {
			c.JSON(200, RespError{false, "图片不存在"})
			return
		}
		data = url
		coach := model.Coach{Id: coachId, Identity: url}
		_, err = session.Id(coachId).Cols("Identity").Update(&coach)
	case 3:
		url := config.CoachLisencePrex + upImage.Key
		if !HasPicture("coachlisence", upImage.Key) {
			c.JSON(200, RespError{false, "图片不存在"})
			return
		}
		data = url
		coach := model.Coach{Id: coachId, CoachLisence: url}
		_, err = session.Id(coachId).Cols("CoachLisence").Update(&coach)
	case 4:
		url := config.DirveLisencePrex + upImage.Key
		if !HasPicture("drivelisence", upImage.Key) {
			c.JSON(200, RespError{false, "图片不存在"})
			return
		}
		data = url
		coach := model.Coach{Id: coachId, DriveLisence: url}
		_, err = session.Id(coachId).Cols("DriveLisence").Update(&coach)
	case 5:
		url := config.SitePicturePrex + upImage.Key
		if !HasPicture("site", upImage.Key) {
			c.JSON(200, RespError{false, "图片不存在"})
			return
		}
		picture = model.Picture{
			UserId:    id,
			Url:       url,
			ImageType: upImage.ImageType}
	case 6:
		url := config.CarPicturPrex + upImage.Key
		if !HasPicture("carpicture", upImage.Key) {
			c.JSON(200, RespError{false, "图片不存在"})
			return
		}
		picture = model.Picture{
			UserId:    id,
			Url:       url,
			ImageType: upImage.ImageType}
	}
	if imageType == 5 || imageType == 6 {
		_, err = session.Insert(&picture)
		picture.UserId = 0
		data = picture
	}
	if err != nil {
		LLog("session.Insert(&picture) in PictureCallback()", err)
		session.Rollback()
		Errors(c)

		return
	}
	err = session.Commit()
	if err != nil {
		LLog("session.Commit() in PictureCallback()", err)
		session.Rollback()
		Errors(c)
		c.JSON(200, 4)
		return
	}
	sess.Delete("upImage")
	c.JSON(200, RespData{true, data})
}

// 获取各类型的图片
func GetPictures(c *macaron.Context, sess session.Store) {
	imageType := c.QueryInt("imageType")
	if imageType < 5 || imageType > 6 {
		c.JSON(200, RespError{false, "没有此类型的图片"})
		return
	}
	id := getuserId(c, sess)
	if id == 0 {
		return
	}
	picture := model.Picture{UserId: id, ImageType: imageType}
	pictures := make([]model.Picture, 0)
	err := engine.Find(&pictures, picture)
	if err != nil {
		LLog("engine.Find(&pictures, picture) in GetPictures()", err)
		Errors(c)
		return
	}
	c.JSON(200, RespPictures{true, pictures})
}

// 删除图片
func DeletePicture(c *macaron.Context, sess session.Store) {
	pictureId := c.ParamsInt64("id")
	if pictureId <= 0 {
		c.JSON(200, RespError{false, "参数错误"})
		return
	}
	coachId := getuserId(c, sess)
	if coachId == 0 {
		return
	}
	picture := model.Picture{Id: pictureId, UserId: coachId}
	ok, err := engine.Get(&picture)
	if err != nil {
		LLog("engine.Get(&picture) in DeletePicture", err)
		Errors(c)
		return
	}
	if !ok {
		c.JSON(200, RespError{false, "图片不存在"})
		return
	}
	if picture.ImageType != 5 && picture.ImageType != 6 {
		c.JSON(200, RespError{false, "无法删除此类型的图片"})
		return
	}
	// 处理数据
	kvs := make(map[string]interface{})
	kvs["picture"] = picture
	commitWithDB(c, kvs, func(c *macaron.Context, kvs map[string]interface{}, s *xorm.Session) interface{} {
		v, _ := kvs["picture"]
		picture := v.(model.Picture)
		// commitWithDB(c, picture, func(c *macaron.Context, v interface{}, s *xorm.Session) interface{} {
		// 	picture := v.(model.Picture)
		num, err := s.Delete(&picture)
		if err != nil {
			LLog("s.Delete(&picture) in DeletePicture", err)
			Errors(c)
			return nil
		}
		if num == 0 {
			c.JSON(200, RespError{false, "删除失败"})
			return nil
		}
		// 删除七牛上的图片
		// 图片类型
		var bucket kodo.Bucket
		// picture.Url
		strs := strings.Split(picture.Url, "/")
		key := config.Env + "/" + strs[len(strs)-1]
		client := getClient()
		ctx := context.Background()
		if picture.ImageType == 5 {
			bucket = client.Bucket("site")
			if !HasPicture("site", key) {
				c.JSON(200, RespError{false, "此图片不存在"})
				return nil
			}
		}
		if picture.ImageType == 6 {
			bucket = client.Bucket("carpicture")
			if !HasPicture("carpicture", key) {
				c.JSON(200, RespError{false, "此图片不存在"})
				return nil
			}
		}
		err = bucket.Delete(ctx, key)
		if err != nil {
			LLog("bucket.Delete(ctx, key) in DeletePicture", err)
			Errors(c)
			return nil
		}
		return RespSuccese{true}
	})
}

// 获取七牛客服端
func getClient() *kodo.Client {
	zone := 0
	return kodo.New(zone, nil)
}

// 检查图片是否存在
func HasPicture(bucket, key string) bool {
	client := getClient()
	kodobucket := client.Bucket(bucket)
	ctx := context.Background()
	_, err := kodobucket.Stat(ctx, key) // 看看空间中是否存在某个文件，其属性是什么
	if err != nil {
		LLog("bucket.Stat(ctx, key)  in HasPicture()", err)
		return false
	}
	return true
}
