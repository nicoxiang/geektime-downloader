package response

// V1VideoPlayAuthResponse ...
type V1VideoPlayAuthResponse struct {
	Code int `json:"code"`
	Data struct {
		PlayAuth string `json:"play_auth"`
		VID      string `json:"vid"`
	} `json:"data"`
}