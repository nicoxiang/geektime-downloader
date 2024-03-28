package response

type V1EnterpriseProductInfoResponse struct {
	Code int `json:"code"`
	Data struct {
		ID          int    `json:"id"`
		SKU         int    `json:"sku"`
		Title       string `json:"title"`
		SubTitle    string `json:"sub_title"`
		ProductType string `json:"product_type"`
		ColumnType  int    `json:"column_type"`
		CourseType  int    `json:"course_type"`
		UpdateFreq  string `json:"update_frequency"`
		Author      struct {
			Name      string `json:"name"`
			Intro     string `json:"intro"`
			Info      string `json:"info"`
			Avatar    string `json:"avatar"`
			BriefHTML string `json:"brief_html"`
			Brief     string `json:"brief"`
		} `json:"author"`
		Cover struct {
			Square            string `json:"square"`
			Rectangle         string `json:"rectangle"`
			Horizontal        string `json:"horizontal"`
			LectureHorizontal string `json:"lecture_horizontal"`
			LearnHorizontal   string `json:"learn_horizontal"`
			Transparent       string `json:"transparent"`
			Color             string `json:"color"`
			Cover             string `json:"cover"`
			RectCover         string `json:"rect_cover"`
			Ratio1            string `json:"ratio_1"`
			Ratio4            string `json:"ratio_4"`
			Ratio16           string `json:"ratio_16"`
			CoverID           int    `json:"cover_id"`
			CoverStatus       int    `json:"cover_status"`
		} `json:"cover"`
		TeachTypeList     []int    `json:"teach_type_list"`
		TeachTypeNameList []string `json:"teach_type_name_list"`
		Article           struct {
			Count            int    `json:"count"`
			CountReq         int    `json:"count_req"`
			CountPub         int    `json:"count_pub"`
			FirstArticleID   string `json:"first_article_id"`
			TotalLength      int    `json:"total_length"`
			TotalTimeStr     string `json:"total_time_str"`
			TotalTimeHourStr string `json:"total_time_hour_str"`
		} `json:"article"`
		SEO struct {
			Keywords []string `json:"keywords"`
		} `json:"seo"`
		Category struct {
			CategoryID int    `json:"category_id"`
			Name       string `json:"name"`
			PID        int    `json:"pid"`
		} `json:"category"`
		Path struct {
			Desc     string `json:"desc"`
			DescHTML string `json:"desc_html"`
		} `json:"path"`
		DL struct {
			Article struct {
				ArticleID       string `json:"article_id"`
				Duration        string `json:"duration"`
				Hot             int    `json:"hot"`
				CouldPreview    bool   `json:"could_preview"`
				DurationSeconds int    `json:"duration_seconds"`
			} `json:"article"`
			CollectionIDs interface{} `json:"collection_ids"`
		} `json:"dl"`
		Share struct {
			PicURL  string `json:"pic_url"`
			Title   string `json:"title"`
			PicName string `json:"pic_name"`
			Content string `json:"content"`
		} `json:"share"`
		IsFinish      bool   `json:"is_finish"`
		Unit          string `json:"unit"`
		BannerCover   string `json:"banner_cover"`
		CatalogPicURL string `json:"catalog_pic_url"`
		Extra         struct {
			Fav struct {
				HasDone    bool `json:"has_done"`
				TotalCount int  `json:"total_count"`
				FavID      int  `json:"fav_id"`
				FavType    int  `json:"fav_type"`
			} `json:"fav"`
			IsSVIP     bool `json:"is_svip"`
			IsMyCourse bool `json:"is_my_course"`
			Rate       struct {
				ArticleCount    int    `json:"article_count"`
				ArticleCountReq int    `json:"article_count_req"`
				IsFinished      bool   `json:"is_finished"`
				RatePercent     int    `json:"rate_percent"`
				VideoSeconds    int    `json:"video_seconds"`
				LastArticleID   string `json:"last_article_id"`
				LastChapterID   int    `json:"last_chapter_id"`
				HasLearn        bool   `json:"has_learn"`
			} `json:"rate"`
			StudyCount int `json:"study_count"`
			Modules    []struct {
				Name    string `json:"name"`
				IsTop   bool   `json:"is_top"`
				Title   string `json:"title"`
				Type    string `json:"type"`
				Content string `json:"content"`
			} `json:"modules"`
			TplType        int           `json:"tpl_type"`
			CollectionType int           `json:"collection_type"`
			WithVideo      bool          `json:"with_video"`
			PIDs           []interface{} `json:"pids"`
			Labels         []interface{} `json:"labels"`
			CategoryIDs    []interface{} `json:"category_ids"`
			Group          struct {
				Title       string `json:"title"`
				Description string `json:"description"`
				StartTime   int    `json:"start_time"`
				EndTime     int    `json:"end_time"`
				QRCodeShow  bool   `json:"qrcode_show"`
				QRCodeURL   string `json:"qrcode_url"`
			} `json:"group"`
			VIP struct {
				Show    bool `json:"show"`
				EndTime int  `json:"end_time"`
			} `json:"vip"`
			CourseStatus   int `json:"course_status"`
			CID            int `json:"cid"`
			RelatedVIPSkus []struct {
				ColumnTitle    string `json:"column_title"`
				DisplayType    int    `json:"display_type"`
				EsPrice        int    `json:"es_price"`
				EsSaleMaxLimit int    `json:"es_sale_max_limit"`
				EsSaleMinLimit int    `json:"es_sale_min_limit"`
				SKU            int    `json:"sku"`
				Status         int    `json:"status"`
				VIPDays        int    `json:"vip_days"`
				VIPTitle       string `json:"vip_title"`
			} `json:"related_vip_skus"`
		} `json:"extra"`
		Intro              string `json:"intro"`
		IntroHTML          string `json:"intro_html"`
		BgColor            string `json:"bgcolor"`
		IsIncludePreview   bool   `json:"is_include_preview"`
		ShowChapter        bool   `json:"show_chapter"`
		DisplayType        int    `json:"display_type"`
		IntroBGStyle       int    `json:"intro_bg_style"`
		Sort               int    `json:"sort"`
		CTime              int    `json:"ctime"`
		SalePrice          int    `json:"sale_price"`
		SaleLimit          int    `json:"sale_limit"`
		Status             int    `json:"status"`
		IsJoinSVIP         int    `json:"is_join_svip"`
		IsJoinColumnVIP    int    `json:"is_join_column_vip"`
		IsJoinCVIP         int    `json:"is_join_cvip"`
		NeedGraduate       int    `json:"need_graduate"`
		AuthorSignatureURL string `json:"author_signature_url"`
		IsFreebie          int    `json:"is_freebie"`
		IsDtai             int    `json:"is_dtai"`
	} `json:"data"`
	Error struct {
	} `json:"error"`
	Extra struct {
		Cost      float64 `json:"cost"`
		RequestID string  `json:"request-id"`
	} `json:"extra"`
}
