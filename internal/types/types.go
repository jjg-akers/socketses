package types

type Permission struct {
	Key         Key         `json:"key"`
	CacheUpdate CacheUpdate `json:"cacheUpdate"`
}

type Key struct {
	Type string `json:"type"`
	Key  string `json:"key"`
}

type CacheUpdate struct {
	Key        string            `json:"key"`
	Value      map[string]string `json:"value"`
	Expiration int64             `json:"expiration"`
}

type SetRequest struct {
	App    string `json:"app"`
	Device string `json:"deviceID"`
	Name   string `json:"name"`
	Value  string `json:"value"`
}
