package response

// V3ColumnInfoResponse ...
type V3ColumnInfoResponse struct {
	Code int `json:"code"`
	Data struct {
		ID               int    `json:"id"`
		Type             string `json:"type"`
		// IsCore           bool   `json:"is_core"`
		IsVideo          bool   `json:"is_video"`
		// IsAudio          bool   `json:"is_audio"`
		// IsSaleProduct    bool   `json:"is_sale_product"`
		// Ucode            string `json:"ucode"`
		// IsFinish         bool   `json:"is_finish"`
		// TotalLength      int    `json:"total_length"`
		// NavID            int    `json:"nav_id"`
		// NpsMin           int    `json:"nps_min"`
		// UpdateFrequency  string `json:"update_frequency"`
		// LastChapterID    int    `json:"last_chapter_id"`
		// BeginTime        int    `json:"begin_time"`
		// EndTime          int    `json:"end_time"`
		// IsPreorder       bool   `json:"is_preorder"`
		// Bgcolor          string `json:"bgcolor"`
		// IsIncludePreview bool   `json:"is_include_preview"`
		// ShowChapter      bool   `json:"show_chapter"`
		// Utime            int    `json:"utime"`
		// IsDailylesson    bool   `json:"is_dailylesson"`
		// IsOpencourse     bool   `json:"is_opencourse"`
		Title            string `json:"title"`
		// Subtitle         string `json:"subtitle"`
		// Ctime            int    `json:"ctime"`
		// Unit             string `json:"unit"`
		// Cover            struct {
		// 	Square      string `json:"square"`
		// 	Rectangle   string `json:"rectangle"`
		// 	Horizontal  string `json:"horizontal"`
		// 	Transparent string `json:"transparent"`
		// 	Color       string `json:"color"`
		// } `json:"cover"`
		// Author struct {
		// 	Name      string `json:"name"`
		// 	Intro     string `json:"intro"`
		// 	Avatar    string `json:"avatar"`
		// 	BriefHTML string `json:"brief_html"`
		// 	Brief     string `json:"brief"`
		// } `json:"author"`
		// Price struct {
		// 	Market       int `json:"market"`
		// 	Sale         int `json:"sale"`
		// 	SaleType     int `json:"sale_type"`
		// 	PromoEndTime int `json:"promo_end_time"`
		// 	StartTime    int `json:"start_time"`
		// 	EndTime      int `json:"end_time"`
		// } `json:"price"`
		// Path struct {
		// 	Desc     string `json:"desc"`
		// 	DescHTML string `json:"desc_html"`
		// } `json:"path"`
		// Article struct {
		// 	ID       int `json:"id"`
		// 	Count    int `json:"count"`
		// 	CountReq int `json:"count_req"`
		// 	CountPub int `json:"count_pub"`
		// } `json:"article"`
		// IsOnborad  bool `json:"is_onborad"`
		// IsSale     bool `json:"is_sale"`
		// IsGroupbuy bool `json:"is_groupbuy"`
		// Seo        struct {
		// 	Keywords []string `json:"keywords"`
		// } `json:"seo"`
		// Share struct {
		// 	Title   string `json:"title"`
		// 	Content string `json:"content"`
		// 	Cover   string `json:"cover"`
		// 	Poster  string `json:"poster"`
		// } `json:"share"`
		// StudyService []int `json:"study_service"`
		// Column       struct {
		// 	CatalogPicURL string `json:"catalog_pic_url"`
		// 	Ranks         []struct {
		// 		ID    int    `json:"id"`
		// 		Name  string `json:"name"`
		// 		Score int    `json:"score"`
		// 	} `json:"ranks"`
		// 	HotComments []struct {
		// 		HadLiked     bool   `json:"had_liked"`
		// 		ProductID    int    `json:"product_id"`
		// 		CommentIsTop bool   `json:"comment_is_top"`
		// 		ID           int    `json:"id"`
		// 		UserHeader   string `json:"user_header"`
		// 		CommentCtime int    `json:"comment_ctime"`
		// 		CanDelete    bool   `json:"can_delete"`
		// 		UserName     string `json:"user_name"`
		// 		IsPvip       bool   `json:"is_pvip"`
		// 		RaceModel    int    `json:"race_model"`
		// 		ProductType  string `json:"product_type"`
		// 		Replies      []struct {
		// 			Ctime        int    `json:"ctime"`
		// 			UserNameReal string `json:"user_name_real"`
		// 			UID          int    `json:"uid"`
		// 			ID           int    `json:"id"`
		// 			Content      string `json:"content"`
		// 			Utype        int    `json:"utype"`
		// 			CommentID    int    `json:"comment_id"`
		// 			UserName     string `json:"user_name"`
		// 			IPAddress    string `json:"ip_address"`
		// 		} `json:"replies"`
		// 		UID            int    `json:"uid"`
		// 		Ucode          string `json:"ucode"`
		// 		Score          int    `json:"score"`
		// 		CommentContent string `json:"comment_content"`
		// 		LikeCount      int    `json:"like_count"`
		// 		Aid            int    `json:"aid"`
		// 		IPAddress      string `json:"ip_address"`
		// 	} `json:"hot_comments"`
		// 	HotLines []struct {
		// 		ProductID   int    `json:"product_id"`
		// 		ProductType string `json:"product_type"`
		// 		Aid         int    `json:"aid"`
		// 		Tips        string `json:"tips"`
		// 		Note        string `json:"note"`
		// 		From        string `json:"from"`
		// 		UlineID     int    `json:"uline_id"`
		// 		UserCount   int    `json:"user_count"`
		// 	} `json:"hot_lines"`
		// 	IntroBgStyle int  `json:"intro_bg_style"`
		// 	InRank       bool `json:"in_rank"`
		// } `json:"column"`
		Extra struct {
			Sub struct {
				// Count      int  `json:"count"`
				// HadDone    bool `json:"had_done"`
				// CouldOrder bool `json:"could_order"`
				AccessMask int  `json:"access_mask"`
			} `json:"sub"`
			// Fav struct {
			// 	Count   int  `json:"count"`
			// 	HadDone bool `json:"had_done"`
			// } `json:"fav"`
			// Rate struct {
			// 	ArticleCount    int  `json:"article_count"`
			// 	ArticleCountReq int  `json:"article_count_req"`
			// 	IsFinished      bool `json:"is_finished"`
			// 	LastArticleID   int  `json:"last_article_id"`
			// 	RatePercent     int  `json:"rate_percent"`
			// 	VideoSeconds    int  `json:"video_seconds"`
			// } `json:"rate"`
			// Cert struct {
			// 	ID string `json:"id"`
			// } `json:"cert"`
			// Sharesale struct {
			// 	Is         bool   `json:"is"`
			// 	Title      string `json:"title"`
			// 	Data       string `json:"data"`
			// 	IsShareget bool   `json:"is_shareget"`
			// 	Amount     int    `json:"amount"`
			// 	MaxAmount  int    `json:"max_amount"`
			// } `json:"sharesale"`
			// Channel struct {
			// 	Is         bool `json:"is"`
			// 	BackAmount int  `json:"back_amount"`
			// } `json:"channel"`
			// Nps struct {
			// 	Status int    `json:"status"`
			// 	URL    string `json:"url"`
			// } `json:"nps"`
			// Helper []struct {
			// 	Title string `json:"title"`
			// 	Desc  string `json:"desc"`
			// 	Icon  string `json:"icon"`
			// } `json:"helper"`
			// Tab struct {
			// 	Comment bool `json:"comment"`
			// } `json:"tab"`
			// GroupBuy struct {
			// 	SuccessUcount int           `json:"success_ucount"`
			// 	JoinCode      string        `json:"join_code"`
			// 	CouldGroupbuy bool          `json:"could_groupbuy"`
			// 	HadJoin       bool          `json:"had_join"`
			// 	Price         int           `json:"price"`
			// 	List          []interface{} `json:"list"`
			// } `json:"group_buy"`
			// AnyRead struct {
			// 	Total int `json:"total"`
			// 	Count int `json:"count"`
			// } `json:"any_read"`
			// Vip struct {
			// 	Show    bool `json:"show"`
			// 	EndTime int  `json:"end_time"`
			// } `json:"vip"`
			// Modules []struct {
			// 	Name    string `json:"name"`
			// 	Title   string `json:"title"`
			// 	Content string `json:"content"`
			// 	Type    string `json:"type"`
			// 	IsTop   bool   `json:"is_top"`
			// } `json:"modules"`
			// Cid        int `json:"cid"`
			// FirstPromo struct {
			// 	Price     int  `json:"price"`
			// 	CouldJoin bool `json:"could_join"`
			// } `json:"first_promo"`
			// StudyPlan struct {
			// 	ID int `json:"id"`
			// } `json:"study_plan"`
			// FirstAward struct {
			// 	Show          bool   `json:"show"`
			// 	Talks         int    `json:"talks"`
			// 	Reads         int    `json:"reads"`
			// 	Amount        int    `json:"amount"`
			// 	ExpireTime    int    `json:"expire_time"`
			// 	RedirectType  string `json:"redirect_type"`
			// 	RedirectParam string `json:"redirect_param"`
			// } `json:"first_award"`
			// SelectCommentCount int   `json:"select_comment_count"`
			// FirstAids          []int `json:"first_aids"`
			// VipPromo           struct {
			// 	DiscountLevel int `json:"discount_level"`
			// 	DiscountPrice int `json:"discount_price"`
			// 	MinLevel      int `json:"min_level"`
			// 	Rules         []struct {
			// 		Tip   string `json:"tip"`
			// 		Price int    `json:"price"`
			// 		Level int    `json:"level"`
			// 	} `json:"rules"`
			// } `json:"vip_promo"`
		} `json:"extra"`
		// FavQrcode  string `json:"fav_qrcode"`
		// Opencourse struct {
		// 	VideoBg string `json:"video_bg"`
		// 	Ad      struct {
		// 		Cover         string `json:"cover"`
		// 		RedirectType  string `json:"redirect_type"`
		// 		RedirectParam string `json:"redirect_param"`
		// 	} `json:"ad"`
		// 	ArticleFav struct {
		// 		Aid     int  `json:"aid"`
		// 		HadDone bool `json:"had_done"`
		// 		Count   int  `json:"count"`
		// 	} `json:"article_fav"`
		// } `json:"opencourse"`
		// RecommendArticles []struct {
		// 	ID           int    `json:"id"`
		// 	Type         int    `json:"type"`
		// 	Pid          int    `json:"pid"`
		// 	ChapterID    int    `json:"chapter_id"`
		// 	ChapterTitle string `json:"chapter_title"`
		// 	Title        string `json:"title"`
		// 	Subtitle     string `json:"subtitle"`
		// 	ShareTitle   string `json:"share_title"`
		// 	Summary      string `json:"summary"`
		// 	Ctime        int    `json:"ctime"`
		// 	Cover        struct {
		// 		Default string `json:"default"`
		// 	} `json:"cover"`
		// 	Author struct {
		// 		Name   string `json:"name"`
		// 		Avatar string `json:"avatar"`
		// 	} `json:"author"`
		// 	Audio struct {
		// 		Title       string   `json:"title"`
		// 		Dubber      string   `json:"dubber"`
		// 		DownloadURL string   `json:"download_url"`
		// 		Md5         string   `json:"md5"`
		// 		Size        int      `json:"size"`
		// 		Time        string   `json:"time"`
		// 		TimeArr     []string `json:"time_arr"`
		// 		URL         string   `json:"url"`
		// 	} `json:"audio"`
		// 	Video struct {
		// 		ID        string        `json:"id"`
		// 		Duration  int           `json:"duration"`
		// 		Cover     string        `json:"cover"`
		// 		Width     int           `json:"width"`
		// 		Height    int           `json:"height"`
		// 		Size      int           `json:"size"`
		// 		Time      string        `json:"time"`
		// 		Medias    []interface{} `json:"medias"`
		// 		HlsVid    string        `json:"hls_vid"`
		// 		HlsMedias []interface{} `json:"hls_medias"`
		// 		Subtitles []interface{} `json:"subtitles"`
		// 		Tips      []interface{} `json:"tips"`
		// 	} `json:"video"`
		// 	VideoPreview struct {
		// 		Duration int         `json:"duration"`
		// 		Medias   interface{} `json:"medias"`
		// 	} `json:"video_preview"`
		// 	VideoPreviews        []interface{} `json:"video_previews"`
		// 	InlineVideoSubtitles []interface{} `json:"inline_video_subtitles"`
		// 	CouldPreview         bool          `json:"could_preview"`
		// 	VideoCouldPreview    bool          `json:"video_could_preview"`
		// 	CoverHidden          bool          `json:"cover_hidden"`
		// 	Content              string        `json:"content"`
		// 	IsRequired           bool          `json:"is_required"`
		// 	Extra                struct {
		// 		Rate []struct {
		// 			Type           int  `json:"type"`
		// 			CurVersion     int  `json:"cur_version"`
		// 			CurRate        int  `json:"cur_rate"`
		// 			MaxRate        int  `json:"max_rate"`
		// 			TotalRate      int  `json:"total_rate"`
		// 			LearnedSeconds int  `json:"learned_seconds"`
		// 			IsFinished     bool `json:"is_finished"`
		// 		} `json:"rate"`
		// 		RatePercent int  `json:"rate_percent"`
		// 		IsFinished  bool `json:"is_finished"`
		// 		Fav         struct {
		// 			Count   int  `json:"count"`
		// 			HadDone bool `json:"had_done"`
		// 		} `json:"fav"`
		// 		IsUnlocked bool `json:"is_unlocked"`
		// 		Learn      struct {
		// 			Ucount int `json:"ucount"`
		// 		} `json:"learn"`
		// 		FooterCoverData struct {
		// 			ImgURL  string `json:"img_url"`
		// 			MpURL   string `json:"mp_url"`
		// 			LinkURL string `json:"link_url"`
		// 		} `json:"footer_cover_data"`
		// 	} `json:"extra"`
		// 	Score           int    `json:"score"`
		// 	IsVideo         bool   `json:"is_video"`
		// 	PosterWxlite    string `json:"poster_wxlite"`
		// 	HadFreelyread   bool   `json:"had_freelyread"`
		// 	FloatQrcode     string `json:"float_qrcode"`
		// 	FloatAppQrcode  string `json:"float_app_qrcode"`
		// 	FloatQrcodeJump string `json:"float_qrcode_jump"`
		// 	InPvip          int    `json:"in_pvip"`
		// 	CommentCount    int    `json:"comment_count"`
		// 	Cshort          string `json:"cshort"`
		// 	Like            struct {
		// 		Count   int  `json:"count"`
		// 		HadDone bool `json:"had_done"`
		// 	} `json:"like"`
		// 	ReadingTime int    `json:"reading_time"`
		// 	IPAddress   string `json:"ip_address"`
		// } `json:"recommend_articles"`
		// InPvip int `json:"in_pvip"`
	} `json:"data"`
	// Error struct {
	// } `json:"error"`
	// Extra struct {
	// 	Cost      float64 `json:"cost"`
	// 	RequestID string  `json:"request-id"`
	// } `json:"extra"`
}