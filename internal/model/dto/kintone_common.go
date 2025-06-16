package dto

import "jaystar/internal/constant/kintone"

type KintoneWebhookIO struct {
	Id          string                 `json:"id"`
	Type        kintone.WebhookType    `json:"type"`
	App         KintoneWebhookApp      `json:"app"`
	RecordId    string                 `json:"recordId"`
	DeleteBy    KintoneWebhookDeleteBy `json:"deleteBy"`
	DeletedAt   string                 `json:"deletedAt"`
	RecordTitle string                 `json:"recordTitle"`
	Url         string                 `json:"url"`
}

type KintoneWebhookApp struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type KintoneWebhookDeleteBy struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type KintoneApiBaseResponse struct {
	Code    string `json:"code"`
	Id      string `json:"id"`
	Message string `json:"message"`
}

type KintoneApiUpdateBaseResponse struct {
	Id       string `json:"id,omitempty"`
	Revision string `json:"revision"`
}

type KintoneUpdateAppBase struct {
	App string `json:"app"`
}

type KintoneUpdateIdBase struct {
	Id int `json:"id"`
}
