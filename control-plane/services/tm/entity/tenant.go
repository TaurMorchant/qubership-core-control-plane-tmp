package entity

type Tenant struct {
	ObjectId   string `json:"objectId"`
	ExternalId string `json:"externalId"`
	Namespace  string `json:"namespace"`
	Status     string `json:"status"`
}

type WatchApiTenant struct {
	Tenants []*Tenant `json:"tenants"`
}
