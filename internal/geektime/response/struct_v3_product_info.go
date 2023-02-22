package response

// V3ProductInfoResponse ...
type V3ProductInfoResponse struct {
	Code int `json:"code"`
	Data struct {
		Info struct {
			ID        int `json:"id"`
			// Spu       int `json:"spu"`
			// Ctime     int `json:"ctime"`
			// Utime     int `json:"utime"`
			// BeginTime int `json:"begin_time"`
			// EndTime   int `json:"end_time"`
			// Price     struct {
			// 	Market    int `json:"market"`
			// 	Sale      int `json:"sale"`
			// 	SaleType  int `json:"sale_type"`
			// 	StartTime int `json:"start_time"`
			// 	EndTime   int `json:"end_time"`
			// } `json:"price"`
			// IsOnborad     bool   `json:"is_onborad"`
			// IsSale        bool   `json:"is_sale"`
			// IsGroupbuy    bool   `json:"is_groupbuy"`
			// IsPromo       bool   `json:"is_promo"`
			// IsShareget    bool   `json:"is_shareget"`
			// IsSharesale   bool   `json:"is_sharesale"`
			Type          string `json:"type"`
			// IsColumn      bool   `json:"is_column"`
			// IsCore        bool   `json:"is_core"`
			IsVideo       bool   `json:"is_video"`
			IsAudio       bool   `json:"is_audio"`
			IsDailylesson bool   `json:"is_dailylesson"`
			// IsUniversity  bool   `json:"is_university"`
			// IsOpencourse  bool   `json:"is_opencourse"`
			// IsQconp       bool   `json:"is_qconp"`
			// NavID         int    `json:"nav_id"`
			// TimeNotSale   int    `json:"time_not_sale"`
			Title         string `json:"title"`
			// Subtitle      string `json:"subtitle"`
			// Intro         string `json:"intro"`
			// IntroHTML     string `json:"intro_html"`
			// Ucode         string `json:"ucode"`
			// IsFinish      bool   `json:"is_finish"`
			// Author        struct {
			// 	Name      string `json:"name"`
			// 	Intro     string `json:"intro"`
			// 	Info      string `json:"info"`
			// 	Avatar    string `json:"avatar"`
			// 	BriefHTML string `json:"brief_html"`
			// 	Brief     string `json:"brief"`
			// } `json:"author"`
			// Cover struct {
			// 	Square            string `json:"square"`
			// 	Rectangle         string `json:"rectangle"`
			// 	Horizontal        string `json:"horizontal"`
			// 	LectureHorizontal string `json:"lecture_horizontal"`
			// 	LearnHorizontal   string `json:"learn_horizontal"`
			// 	Transparent       string `json:"transparent"`
			// 	Color             string `json:"color"`
			// } `json:"cover"`
			Article struct {
				ID                int    `json:"id"`
				// Count             int    `json:"count"`
				// CountReq          int    `json:"count_req"`
				// CountPub          int    `json:"count_pub"`
				// TotalLength       int    `json:"total_length"`
				// FirstArticleID    int    `json:"first_article_id"`
				// FirstArticleTitle string `json:"first_article_title"`
			} `json:"article"`
			// Seo struct {
			// 	Keywords []interface{} `json:"keywords"`
			// } `json:"seo"`
			// Share struct {
			// 	Title   string `json:"title"`
			// 	Content string `json:"content"`
			// 	Cover   string `json:"cover"`
			// 	Poster  string `json:"poster"`
			// } `json:"share"`
			// Labels     []int  `json:"labels"`
			// Unit       string `json:"unit"`
			ColumnType int    `json:"column_type"`
			// Column     struct {
			// 	Unit             string        `json:"unit"`
			// 	Bgcolor          string        `json:"bgcolor"`
			// 	UpdateFrequency  string        `json:"update_frequency"`
			// 	IsPreorder       bool          `json:"is_preorder"`
			// 	IsFinish         bool          `json:"is_finish"`
			// 	IsIncludePreview bool          `json:"is_include_preview"`
			// 	ShowChapter      bool          `json:"show_chapter"`
			// 	IsSaleProduct    bool          `json:"is_sale_product"`
			// 	StudyService     []interface{} `json:"study_service"`
			// 	Path             struct {
			// 		Desc     string `json:"desc"`
			// 		DescHTML string `json:"desc_html"`
			// 	} `json:"path"`
			// 	IsCamp                   bool        `json:"is_camp"`
			// 	CatalogPicURL            string      `json:"catalog_pic_url"`
			// 	RecommendArticles        interface{} `json:"recommend_articles"`
			// 	RecommendComments        interface{} `json:"recommend_comments"`
			// 	Ranks                    interface{} `json:"ranks"`
			// 	HotComments              interface{} `json:"hot_comments"`
			// 	HotLines                 interface{} `json:"hot_lines"`
			// 	DisplayType              int         `json:"display_type"`
			// 	IntroBgStyle             int         `json:"intro_bg_style"`
			// 	CommentTopAds            string      `json:"comment_top_ads"`
			// 	ArticleFloatQrcodeURL    string      `json:"article_float_qrcode_url"`
			// 	ArticleFloatAppQrcodeURL string      `json:"article_float_app_qrcode_url"`
			// 	ArticleFloatQrcodeJump   string      `json:"article_float_qrcode_jump"`
			// 	InRank                   bool        `json:"in_rank"`
			// } `json:"column"`
			// Dl struct {
			// 	Article struct {
			// 		ID            int    `json:"id"`
			// 		VideoDuration string `json:"video_duration"`
			// 		VideoHot      int    `json:"video_hot"`
			// 		CouldPreview  bool   `json:"could_preview"`
			// 	} `json:"article"`
			// 	TopicIds []interface{} `json:"topic_ids"`
			// } `json:"dl"`
			// University struct {
			// 	TotalHour       int    `json:"total_hour"`
			// 	Term            int    `json:"term"`
			// 	RedirectType    string `json:"redirect_type"`
			// 	RedirectParam   string `json:"redirect_param"`
			// 	WxQrcode        string `json:"wx_qrcode"`
			// 	WxRule          string `json:"wx_rule"`
			// 	ServerStartTime int    `json:"server_start_time"`
			// 	LecturerHCover  string `json:"lecturer_h_cover"`
			// } `json:"university"`
			// Opencourse struct {
			// 	VideoBg string `json:"video_bg"`
			// 	Ad      struct {
			// 		Cover         string `json:"cover"`
			// 		CoverWeb      string `json:"cover_web"`
			// 		RedirectType  string `json:"redirect_type"`
			// 		RedirectParam string `json:"redirect_param"`
			// 	} `json:"ad"`
			// 	ArticleFav struct {
			// 		Aid     int  `json:"aid"`
			// 		HadDone bool `json:"had_done"`
			// 		Count   int  `json:"count"`
			// 	} `json:"article_fav"`
			// 	AuthorHCover string `json:"author_h_cover"`
			// } `json:"opencourse"`
			// Qconp struct {
			// 	TopicID      int    `json:"topic_id"`
			// 	CoverAppoint string `json:"cover_appoint"`
			// 	Article      struct {
			// 		ID            int    `json:"id"`
			// 		Cover         string `json:"cover"`
			// 		VideoDuration string `json:"video_duration"`
			// 		VideoHot      int    `json:"video_hot"`
			// 	} `json:"article"`
			// } `json:"qconp"`
			// FavQrcode string `json:"fav_qrcode"`
			Extra     struct {
				Sub struct {
					Count      int  `json:"count"`
					HadDone    bool `json:"had_done"`
					CouldOrder bool `json:"could_order"`
					AccessMask int  `json:"access_mask"`
				} `json:"sub"`
			// 	Fav struct {
			// 		Count   int  `json:"count"`
			// 		HadDone bool `json:"had_done"`
			// 	} `json:"fav"`
			// 	Rate struct {
			// 		ArticleCount    int  `json:"article_count"`
			// 		ArticleCountReq int  `json:"article_count_req"`
			// 		IsFinished      bool `json:"is_finished"`
			// 		RatePercent     int  `json:"rate_percent"`
			// 		VideoSeconds    int  `json:"video_seconds"`
			// 		LastArticleID   int  `json:"last_article_id"`
			// 		LastChapterID   int  `json:"last_chapter_id"`
			// 		HasLearn        bool `json:"has_learn"`
			// 	} `json:"rate"`
			// 	Cert struct {
			// 		ID   string `json:"id"`
			// 		Type int    `json:"type"`
			// 	} `json:"cert"`
			// 	Nps struct {
			// 		Min    int    `json:"min"`
			// 		Status int    `json:"status"`
			// 		URL    string `json:"url"`
			// 	} `json:"nps"`
			// 	AnyRead struct {
			// 		Total int `json:"total"`
			// 		Count int `json:"count"`
			// 	} `json:"any_read"`
			// 	University struct {
			// 		Status               int           `json:"status"`
			// 		ViewStatus           int           `json:"view_status"`
			// 		ChargeStatus         int           `json:"charge_status"`
			// 		ShareRenewalStatus   int           `json:"share_renewal_status"`
			// 		UnlockedStatus       int           `json:"unlocked_status"`
			// 		UnlockedChapterIds   []interface{} `json:"unlocked_chapter_ids"`
			// 		UnlockedChapterID    int           `json:"unlocked_chapter_id"`
			// 		UnlockedChapterTitle string        `json:"unlocked_chapter_title"`
			// 		UnlockedArticleCount int           `json:"unlocked_article_count"`
			// 		UnlockedNextTime     int           `json:"unlocked_next_time"`
			// 		ExpireTime           int           `json:"expire_time"`
			// 		IsExpired            bool          `json:"is_expired"`
			// 		IsGraduated          bool          `json:"is_graduated"`
			// 		HadSub               bool          `json:"had_sub"`
			// 		Timeline             []interface{} `json:"timeline"`
			// 		HasWxFriend          bool          `json:"has_wx_friend"`
			// 		SubTermTitle         string        `json:"sub_term_title"`
			// 		SubSku               int           `json:"sub_sku"`
			// 	} `json:"university"`
			// 	Vip struct {
			// 		IsYearCard bool `json:"is_year_card"`
			// 		Show       bool `json:"show"`
			// 		EndTime    int  `json:"end_time"`
			// 	} `json:"vip"`
			// 	Appoint struct {
			// 		CouldDo bool `json:"could_do"`
			// 		HadDone bool `json:"had_done"`
			// 		Count   int  `json:"count"`
			// 	} `json:"appoint"`
			// 	GroupBuy struct {
			// 		SuccessUcount int           `json:"success_ucount"`
			// 		JoinCode      string        `json:"join_code"`
			// 		CouldGroupbuy bool          `json:"could_groupbuy"`
			// 		HadJoin       bool          `json:"had_join"`
			// 		Price         int           `json:"price"`
			// 		List          []interface{} `json:"list"`
			// 	} `json:"group_buy"`
			// 	Sharesale struct {
			// 		OriginalPicColor    string `json:"original_pic_color"`
			// 		OriginalPicURL      string `json:"original_pic_url"`
			// 		PromoPicColor       string `json:"promo_pic_color"`
			// 		PromoPicURL         string `json:"promo_pic_url"`
			// 		ShareSalePrice      int    `json:"share_sale_price"`
			// 		ShareSaleGuestPrice int    `json:"share_sale_guest_price"`
			// 	} `json:"sharesale"`
			// 	Promo struct {
			// 		EntTime int `json:"ent_time"`
			// 	} `json:"promo"`
			// 	Channel struct {
			// 		Is         bool `json:"is"`
			// 		BackAmount int  `json:"back_amount"`
			// 	} `json:"channel"`
			// 	FirstPromo struct {
			// 		Price     int  `json:"price"`
			// 		CouldJoin bool `json:"could_join"`
			// 	} `json:"first_promo"`
			// 	CouponPromo struct {
			// 		CouldJoin bool `json:"could_join"`
			// 		Price     int  `json:"price"`
			// 	} `json:"coupon_promo"`
			// 	Helper []interface{} `json:"helper"`
			// 	Tab    struct {
			// 		Comment bool `json:"comment"`
			// 	} `json:"tab"`
			// 	Modules   []interface{} `json:"modules"`
			// 	Cid       int           `json:"cid"`
			// 	FirstAids []interface{} `json:"first_aids"`
			// 	StudyPlan struct {
			// 		ID              int `json:"id"`
			// 		DayNums         int `json:"day_nums"`
			// 		ArticleNums     int `json:"article_nums"`
			// 		LearnedWeekNums int `json:"learned_week_nums"`
			// 		Status          int `json:"status"`
			// 	} `json:"study_plan"`
			// 	CateID   int    `json:"cate_id"`
			// 	CateName string `json:"cate_name"`
			// 	GroupTag struct {
			// 		IsRecommend     bool `json:"is_recommend"`
			// 		IsRecentlyLearn bool `json:"is_recently_learn"`
			// 	} `json:"group_tag"`
			// 	FirstAward struct {
			// 		Show          bool   `json:"show"`
			// 		Talks         int    `json:"talks"`
			// 		Reads         int    `json:"reads"`
			// 		Amount        int    `json:"amount"`
			// 		ExpireTime    int    `json:"expire_time"`
			// 		RedirectType  string `json:"redirect_type"`
			// 		RedirectParam string `json:"redirect_param"`
			// 	} `json:"first_award"`
			// 	VipPromo struct {
			// 		DiscountLevel int         `json:"discount_level"`
			// 		DiscountPrice int         `json:"discount_price"`
			// 		MinLevel      int         `json:"min_level"`
			// 		Rules         interface{} `json:"rules"`
			// 	} `json:"vip_promo"`
			// 	IsTgoTicket bool `json:"is_tgo_ticket"`
			} `json:"extra"`
			// AvailableCoupons interface{} `json:"available_coupons"`
			InPvip           int         `json:"in_pvip"`
		} `json:"info"`
		// Labels []struct {
		// 	ID    int    `json:"id"`
		// 	Name  string `json:"name"`
		// 	Icon  string `json:"icon"`
		// 	Count int    `json:"count"`
		// 	Pid   int    `json:"pid"`
		// 	Sort  int    `json:"sort"`
		// } `json:"labels"`
		// Topics []interface{} `json:"topics"`
	} `json:"data"`
	Error struct {
	} `json:"error"`
	// Extra struct {
	// 	Cost      float64 `json:"cost"`
	// 	RequestID string  `json:"request-id"`
	// } `json:"extra"`
}