package model

type ResultSMS struct {
	StatusCode  string      `json:"statusCode"`
	SMSTemplate TemplateSMS `json:"TemplateSMS"`
}

type TemplateSMS struct {
	DateCreated   string `json:"dateCreated"`
	SmsMessageSid string `json:"smsMessageSid"`
}
