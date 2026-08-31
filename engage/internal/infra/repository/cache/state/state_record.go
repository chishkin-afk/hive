package statecache

type State struct {
	ClientID int64          `json:"client_id"`
	Step     string         `json:"step"`
	Data     map[string]any `json:"data"`
}
