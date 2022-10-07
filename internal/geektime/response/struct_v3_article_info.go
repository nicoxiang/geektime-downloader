package response

// V3ArticleInfoResponse ...
type V3ArticleInfoResponse struct {
	Code int `json:"code"`
	Data struct {
		Info struct {
			ID           int    `json:"id"`
			Type         int    `json:"type"`
			// Pid          int    `json:"pid"`
			// ChapterID    int    `json:"chapter_id"`
			// ChapterTitle string `json:"chapter_title"`
			Title        string `json:"title"`
			// Subtitle     string `json:"subtitle"`
			// ShareTitle   string `json:"share_title"`
			// Summary      string `json:"summary"`
			// Ctime        int    `json:"ctime"`
			// Cover        struct {
			// 	Default string `json:"default"`
			// } `json:"cover"`
			// Author struct {
			// 	Name   string `json:"name"`
			// 	Avatar string `json:"avatar"`
			// } `json:"author"`
			// Audio struct {
			// 	Title       string   `json:"title"`
			// 	Dubber      string   `json:"dubber"`
			// 	DownloadURL string   `json:"download_url"`
			// 	Md5         string   `json:"md5"`
			// 	Size        int      `json:"size"`
			// 	Time        string   `json:"time"`
			// 	TimeArr     []string `json:"time_arr"`
			// 	URL         string   `json:"url"`
			// } `json:"audio"`
			Video struct {
				ID       string `json:"id"`
				// Duration int    `json:"duration"`
				// Cover    string `json:"cover"`
				// Width    int    `json:"width"`
				// Height   int    `json:"height"`
				// Size     int    `json:"size"`
				// Time     string `json:"time"`
				// Medias   []struct {
				// 	Size    int    `json:"size"`
				// 	Quality string `json:"quality"`
				// 	URL     string `json:"url"`
				// } `json:"medias"`
				// HlsVid    string `json:"hls_vid"`
				// HlsMedias []struct {
				// 	Size    int    `json:"size"`
				// 	Quality string `json:"quality"`
				// 	URL     string `json:"url"`
				// } `json:"hls_medias"`
				// Subtitles []interface{} `json:"subtitles"`
				// Tips      []interface{} `json:"tips"`
			} `json:"video"`
			// VideoPreview struct {
			// 	Duration int `json:"duration"`
			// 	Medias   []struct {
			// 		Size    int    `json:"size"`
			// 		Quality string `json:"quality"`
			// 		URL     string `json:"url"`
			// 	} `json:"medias"`
			// } `json:"video_preview"`
			// VideoPreviews        interface{}   `json:"video_previews"`
			// InlineVideoSubtitles []interface{} `json:"inline_video_subtitles"`
			// CouldPreview         bool          `json:"could_preview"`
			// VideoCouldPreview    bool          `json:"video_could_preview"`
			// CoverHidden          bool          `json:"cover_hidden"`
			// IsRequired           bool          `json:"is_required"`
			// Extra                struct {
			// 	Rate []struct {
			// 		Type           int  `json:"type"`
			// 		CurVersion     int  `json:"cur_version"`
			// 		CurRate        int  `json:"cur_rate"`
			// 		MaxRate        int  `json:"max_rate"`
			// 		TotalRate      int  `json:"total_rate"`
			// 		LearnedSeconds int  `json:"learned_seconds"`
			// 		IsFinished     bool `json:"is_finished"`
			// 	} `json:"rate"`
			// 	RatePercent int  `json:"rate_percent"`
			// 	IsFinished  bool `json:"is_finished"`
			// 	Fav         struct {
			// 		Count   int  `json:"count"`
			// 		HadDone bool `json:"had_done"`
			// 	} `json:"fav"`
			// 	IsUnlocked bool `json:"is_unlocked"`
			// 	Learn      struct {
			// 		Ucount int `json:"ucount"`
			// 	} `json:"learn"`
			// 	FooterCoverData struct {
			// 		ImgURL  string `json:"img_url"`
			// 		MpURL   string `json:"mp_url"`
			// 		LinkURL string `json:"link_url"`
			// 	} `json:"footer_cover_data"`
			// } `json:"extra"`
			// Score           int    `json:"score"`
			IsVideo         bool   `json:"is_video"`
			// PosterWxlite    string `json:"poster_wxlite"`
			// HadFreelyread   bool   `json:"had_freelyread"`
			// FloatQrcode     string `json:"float_qrcode"`
			// FloatAppQrcode  string `json:"float_app_qrcode"`
			// FloatQrcodeJump string `json:"float_qrcode_jump"`
			InPvip          int    `json:"in_pvip"`
			// CommentCount    int    `json:"comment_count"`
			// Cshort          string `json:"cshort"`
			// Like            struct {
			// 	Count   int  `json:"count"`
			// 	HadDone bool `json:"had_done"`
			// } `json:"like"`
			// ReadingTime int           `json:"reading_time"`
			// IPAddress   string        `json:"ip_address"`
			// Content     string        `json:"content"`
			// ContentMd   string        `json:"content_md"`
			// Attachments []interface{} `json:"attachments"`
		} `json:"info"`
		Product struct {
			// ID         int    `json:"id"`
			// Title      string `json:"title"`
			// University struct {
			// 	RedirectType  string `json:"redirect_type"`
			// 	RedirectParam string `json:"redirect_param"`
			// } `json:"university"`
			Extra struct {
				Sub struct {
					// HadDone    bool `json:"had_done"`
					AccessMask int  `json:"access_mask"`
				} `json:"sub"`
			} `json:"extra"`
			Type string `json:"type"`
		} `json:"product"`
		// FreeGet    bool `json:"free_get"`
		// IsFullText bool `json:"is_full_text"`
	} `json:"data"`
	Error struct {
	} `json:"error"`
	// Extra struct {
	// 	Cost      float64 `json:"cost"`
	// 	RequestID string  `json:"request-id"`
	// } `json:"extra"`
}