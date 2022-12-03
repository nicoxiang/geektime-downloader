package response

// V1ArticleResponse ...
type V1ArticleResponse struct {
	// Error []interface{} `json:"error"`
	// Extra []interface{} `json:"extra"`
	Data struct {
		// TextReadVersion int           `json:"text_read_version"`
		// AudioSize       int           `json:"audio_size"`
		// ArticleCover    string        `json:"article_cover"`
		// Subtitles       []interface{} `json:"subtitles"`
		// ProductType     string        `json:"product_type"`
		// AudioDubber     string        `json:"audio_dubber"`
		// IsFinished      bool          `json:"is_finished"`
		// Like            struct {
		// 	HadDone bool `json:"had_done"`
		// 	Count   int  `json:"count"`
		// } `json:"like"`
		// AudioTime string `json:"audio_time"`
		// Share     struct {
		// 	Content string `json:"content"`
		// 	Title   string `json:"title"`
		// 	Poster  string `json:"poster"`
		// 	Cover   string `json:"cover"`
		// } `json:"share"`
		ArticleContent string `json:"article_content"`
		// FloatQrcode        string `json:"float_qrcode"`
		// ArticleCoverHidden bool   `json:"article_cover_hidden"`
		// IsRequired         bool   `json:"is_required"`
		// Score              string `json:"score"`
		// LikeCount          int    `json:"like_count"`
		// ArticleSubtitle    string `json:"article_subtitle"`
		// VideoTime          string `json:"video_time"`
		// HadViewed          bool   `json:"had_viewed"`
		ArticleTitle string `json:"article_title"`
		// ColumnBgcolor      string `json:"column_bgcolor"`
		// OfflinePackage     string `json:"offline_package"`
		// AudioTitle         string `json:"audio_title"`
		// AudioTimeArr       struct {
		// 	M string `json:"m"`
		// 	S string `json:"s"`
		// 	H string `json:"h"`
		// } `json:"audio_time_arr"`
		// TextReadPercent      int           `json:"text_read_percent"`
		// Cid                  int           `json:"cid"`
		// ArticleCshort        string        `json:"article_cshort"`
		// VideoWidth           int           `json:"video_width"`
		// ColumnCouldSub       bool          `json:"column_could_sub"`
		// VideoID              string        `json:"video_id"`
		// Sku                  string        `json:"sku"`
		// VideoCover           string        `json:"video_cover"`
		// AuthorName           string        `json:"author_name"`
		// ColumnIsOnboard      bool          `json:"column_is_onboard"`
		InlineVideoSubtitles []struct {
			VideoURL          string `json:"video_url"`
			VideoFreeURL      string `json:"video_free_url"`
			VideoVid          string `json:"video_vid"`
			VideoSubtitle     string `json:"video_subtitle"`
			VideoFreeVid      string `json:"video_free_vid"`
			VideoFreeSubtitle string `json:"video_free_subtitle"`
		} `json:"inline_video_subtitles"`
		// AudioURL             string        `json:"audio_url"`
		// ChapterID            string        `json:"chapter_id"`
		// ColumnHadSub         bool          `json:"column_had_sub"`
		// ColumnCover          string        `json:"column_cover"`
		// Neighbors            struct {
		// 	Left  []interface{} `json:"left"`
		// 	Right struct {
		// 		ArticleTitle string `json:"article_title"`
		// 		ID           int    `json:"id"`
		// 	} `json:"right"`
		// } `json:"neighbors"`
		// RatePercent     int `json:"rate_percent"`
		// FooterCoverData struct {
		// 	ImgURL  string `json:"img_url"`
		// 	LinkURL string `json:"link_url"`
		// 	MpURL   string `json:"mp_url"`
		// } `json:"footer_cover_data"`
		// FloatAppQrcode     string `json:"float_app_qrcode"`
		// ColumnIsExperience bool   `json:"column_is_experience"`
		// Rate               struct {
		// 	Num1 struct {
		// 		CurVersion     int  `json:"cur_version"`
		// 		MaxRate        int  `json:"max_rate"`
		// 		CurRate        int  `json:"cur_rate"`
		// 		IsFinished     bool `json:"is_finished"`
		// 		TotalRate      int  `json:"total_rate"`
		// 		LearnedSeconds int  `json:"learned_seconds"`
		// 	} `json:"1"`
		// 	Num2 struct {
		// 		CurVersion     int  `json:"cur_version"`
		// 		MaxRate        int  `json:"max_rate"`
		// 		CurRate        int  `json:"cur_rate"`
		// 		IsFinished     bool `json:"is_finished"`
		// 		TotalRate      int  `json:"total_rate"`
		// 		LearnedSeconds int  `json:"learned_seconds"`
		// 	} `json:"2"`
		// 	Num3 struct {
		// 		CurVersion     int  `json:"cur_version"`
		// 		MaxRate        int  `json:"max_rate"`
		// 		CurRate        int  `json:"cur_rate"`
		// 		IsFinished     bool `json:"is_finished"`
		// 		TotalRate      int  `json:"total_rate"`
		// 		LearnedSeconds int  `json:"learned_seconds"`
		// 	} `json:"3"`
		// } `json:"rate"`
		// ProductID           int    `json:"product_id"`
		// HadLiked            bool   `json:"had_liked"`
		// ID                  int    `json:"id"`
		// FreeGet             bool   `json:"free_get"`
		// IsVideoPreview      bool   `json:"is_video_preview"`
		// ArticleSummary      string `json:"article_summary"`
		// ColumnSaleType      int    `json:"column_sale_type"`
		// FloatQrcodeJump     string `json:"float_qrcode_jump"`
		// ColumnID            int    `json:"column_id"`
		// IPAddress           string `json:"ip_address"`
		// AudioMd5            string `json:"audio_md5"`
		// ArticleCouldPreview bool   `json:"article_could_preview"`
		// ArticleSharetitle   string `json:"article_sharetitle"`
		// ArticlePosterWxlite string `json:"article_poster_wxlite"`
		// ArticleFeatures     int    `json:"article_features"`
		// CommentCount        int    `json:"comment_count"`
		// VideoSize           int    `json:"video_size"`
		// Offline             struct {
		// 	Size        int    `json:"size"`
		// 	FileName    string `json:"file_name"`
		// 	DownloadURL string `json:"download_url"`
		// } `json:"offline"`
		// ReadingTime      int           `json:"reading_time"`
		// HlsVideos        []interface{} `json:"hls_videos"`
		// InPvip           int           `json:"in_pvip"`
		AudioDownloadURL string `json:"audio_download_url"`
		// ArticleCtime     int           `json:"article_ctime"`
		// VideoHeight      int           `json:"video_height"`
	} `json:"data"`
	Code int `json:"code"`
}
