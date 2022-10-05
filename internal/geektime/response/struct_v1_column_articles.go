package response

// V1ColumnArticlesResponse ...
type V1ColumnArticlesResponse struct {
	// Error []interface{} `json:"error"`
	// Extra []interface{} `json:"extra"`
	Data  struct {
		List []struct {
			// ArticleSharetitle string        `json:"article_sharetitle,omitempty"`
			// AudioSize         int           `json:"audio_size,omitempty"`
			// ArticleCover      string        `json:"article_cover"`
			// Subtitles         []interface{} `json:"subtitles"`
			// AudioURL          string        `json:"audio_url,omitempty"`
			// ChapterID         string        `json:"chapter_id"`
			// ColumnHadSub      bool          `json:"column_had_sub"`
			// ReadingTime       int           `json:"reading_time"`
			// IsFinished        bool          `json:"is_finished"`
			// AudioTime         string        `json:"audio_time,omitempty"`
			// RatePercent       int           `json:"rate_percent"`
			// ColumnSku         int           `json:"column_sku"`
			// IsRequired        bool          `json:"is_required"`
			// Rate              struct {
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
			// Score            int64  `json:"score"`
			// ArticleSubtitle  string `json:"article_subtitle"`
			// AudioDownloadURL string `json:"audio_download_url,omitempty"`
			ID               int    `json:"id"`
			// HadViewed        bool   `json:"had_viewed"`
			ArticleTitle     string `json:"article_title"`
			// ColumnBgcolor    string `json:"column_bgcolor,omitempty"`
			// IsVideoPreview   bool   `json:"is_video_preview"`
			// ArticleSummary   string `json:"article_summary"`
			// ColumnID         int    `json:"column_id,omitempty"`
			// AudioTitle       string `json:"audio_title,omitempty"`
			// AudioMd5         string `json:"audio_md5,omitempty"`
			// IPAddress        string `json:"ip_address"`
			// AuthorName       string `json:"author_name,omitempty"`
			// AuthorIntro      string `json:"author_intro,omitempty"`
			// Offline          struct {
			// 	FileName    string `json:"file_name"`
			// 	DownloadURL string `json:"download_url"`
			// } `json:"offline"`
			// ColumnCover  string `json:"column_cover,omitempty"`
			// AudioDubber  string `json:"audio_dubber,omitempty"`
			// AudioTimeArr struct {
			// 	M string `json:"m"`
			// 	S string `json:"s"`
			// 	H string `json:"h"`
			// } `json:"audio_time_arr,omitempty"`
			// ArticleCouldPreview bool `json:"article_could_preview"`
			// ArticleCtime        int  `json:"article_ctime"`
			// IncludeAudio        bool `json:"include_audio"`
		} `json:"list"`
		Page struct {
			Count int  `json:"count"`
			More  bool `json:"more"`
		} `json:"page"`
	} `json:"data"`
	Code int `json:"code"`
}