package response

// V3VideoPlayAuthResponse ...
type V3VideoPlayAuthResponse struct {
	Code int `json:"code"`
	Data struct {
		PlayAuth string `json:"play_auth"`
	} `json:"data"`
}