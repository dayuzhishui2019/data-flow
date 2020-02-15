package base

type Related struct {
	RelatedType string `json:"RelatedType"` // 关联类型

	RelatedID string `json:"RelatedID"` //当关联多个信息时，用“,”号隔开
}

type RelatedList struct {
	RelatedObject []*Related `json:"RelatedObject"`
}
