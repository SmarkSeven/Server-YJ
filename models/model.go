package model

// 容联云API返回结果
type VerifycodeResult struct {
	Success bool `json:"success"`
	Result  struct {
		Templatesms struct {
			Datecreated   string `json:"dateCreated"`
			Smsmessagesid string `json:"smsMessageSid"`
		} `json:"templateSMS"`
		Statuscode string `json:"statusCode"`
		Verifycode string `json:"verifyCode"`
	} `json:"result"`
}

// type Coach struct {
// 	Id        int64  `json:"id"`
// 	TeachType int    `json:"teachType"`
// 	Name      string `json:"name"`
// 	Phone     string `json:"phone"`
// }

type VerifyCodeTemp struct {
	Code     string `json:"code"`
	Phone    string `json:"phone"`
	DeadLine string `json:"deadLine"`
}

type Cookie struct {
	Name       string      `json:"Name"`
	Value      string      `json:"Value"`
	Path       string      `json:"Path"`
	Domain     string      `json:"Domain"`
	Expires    string      `json:"Expires"`
	Rawexpires string      `json:"RawExpires"`
	Maxage     int         `json:"MaxAge"`
	Secure     bool        `json:"Secure"`
	Httponly   bool        `json:"HttpOnly"`
	Raw        string      `json:"Raw"`
	Unparsed   interface{} `json:"Unparsed"`
}

// 驾校
type DriveSchool struct {
	Id     int64  `json:"id"`
	Name   string `json:"name"`
	Pinyin string `json:"pinyin"`
	ProvId string `json:"provId"`
	CityId string `json:"cityId"`
}

//返回结果 1-成功，2-错误的cookie，3-参数错误，4-无效的cookie，5-访问时间问题，6－系统错误，7-其它
type Result struct {
	Code      int         `json:"code"`
	ErrorInfo string      `json:"errorInfo"`
	Data      interface{} `json:"data"`
}
