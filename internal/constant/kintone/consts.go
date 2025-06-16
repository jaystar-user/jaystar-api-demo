package kintone

// API URL path
const (
	RecordsPath = "/k/v1/records.json" // 批量記錄
	RecordPath  = "/k/v1/record.json"  // 單筆記錄
)

// API 常用參數名稱
const (
	// header
	HeaderUserAuthorization = "X-Cybozu-Authorization"

	// querystring/body/response
	QueryApp   = "app"
	QueryQuery = "query"
	QueryField = "field[%d]"
	TotalCount = "totalCount"
)

// API 「query」參數常用 keyword
const (
	QueryQueryOrderBy = "order by"
	QueryQueryLimit   = "limit"
	QueryQueryOffset  = "offset"
)

// Webhook 相關
const (
	WebhookClientIP  = "103.79.14.86"
	WebhookUserAgent = "kintone-Webhook/0.1"
)

// API 限制
const (
	BatchInsertRecordsMaxLimit = 100
)
