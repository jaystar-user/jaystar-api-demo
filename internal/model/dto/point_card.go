package dto

type KintoneWebhookPointCardIO struct {
	KintoneWebhookIO
	Record PointCardRecord `json:"record"`
}
