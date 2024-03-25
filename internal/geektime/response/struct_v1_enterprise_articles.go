package response

type V1EnterpriseArticlesResponse struct {
	Code int `json:"code"`
	Data struct {
		List []struct {
			ID          int    `json:"id"`
			Title       string `json:"title"`
			Count       int    `json:"count"`
			Score       int    `json:"score"`
			IsLast      bool   `json:"is_last"`
			ArticleList []struct {
				ID               string `json:"id"`
				Time             string `json:"time"`
				Type             string `json:"type"`
				FavoriteID       int    `json:"favorite_id"`
				DiscussionNumber int    `json:"discussion_number"`
				ColumnTitle      string `json:"column_title"`
				Rights           bool   `json:"rights"`
				Show             bool   `json:"show"`
				RichType         int    `json:"rich_type"`
				PID              int    `json:"pid"`
				SKU              int    `json:"sku"`
				Action           string `json:"action"`
				Score            int    `json:"score"`
				IsRequired       bool   `json:"is_required"`
				URI              string `json:"uri"`
				ColumnType       int    `json:"column_type"`
				EnterpriseID     string `json:"enterprise_id"`
				NodeType         int    `json:"node_type"`
				Published        int    `json:"published"`
				ArtStatus        int    `json:"art_status"`
				SKUStatus        int    `json:"sku_status"`
				IsSell           int    `json:"is_sell"`
				Name             string `json:"name"`
				ProductType      string `json:"product_type"`
				ArticleSource    int    `json:"article_source"`
				ArticleVendorID  int    `json:"article_vendor_id"`
				Author           struct {
					Name   string `json:"name"`
					Avatar string `json:"avatar"`
					Info   string `json:"info"`
					Intro  string `json:"intro"`
				} `json:"author"`
				Article struct {
					ID               string `json:"id"`
					Title            string `json:"title"`
					Content          string `json:"content"`
					ContentMD        string `json:"content_md"`
					CTime            int    `json:"ctime"`
					PosterWxlite     string `json:"poster_wxlite"`
					CoverHidden      int    `json:"cover_hidden"`
					Subtitle         string `json:"subtitle"`
					Summary          string `json:"summary"`
					CouldPreview     bool   `json:"could_preview"`
					BCouldPreview    bool   `json:"b_could_preview"`
					ContentJSON      string `json:"content_json"`
					ContentJSONShort string `json:"content_json_short"`
					InlineVideo      struct {
						Rights  []interface{} `json:"rights"`
						Preview []interface{} `json:"preview"`
					} `json:"inline_video"`
					Cover struct {
						ColumnCover string `json:"column_cover"`
						Default     string `json:"default"`
						CoverID     int    `json:"cover_id"`
						CoverStatus int    `json:"cover_status"`
						SKUCover    struct {
							Ratio16    string `json:"ratio_16"`
							Ratio16URL string `json:"ratio_16_url"`
							Ratio4     string `json:"ratio_4"`
							Ratio4URL  string `json:"ratio_4_url"`
							Ratio1     string `json:"ratio_1"`
							Ratio1URL  string `json:"ratio_1_url"`
							ShowCover  int    `json:"show_cover"`
						} `json:"sku_cover"`
					} `json:"cover"`
					Share struct {
						Title   string `json:"title"`
						Content string `json:"content"`
						Cover   string `json:"cover"`
						Poster  string `json:"poster"`
					} `json:"share"`
					Relation struct {
						PrevID           string `json:"prev_id"`
						PrevChapterTitle string `json:"prev_chapter_title"`
						PrevArticleTitle string `json:"prev_article_title"`
						NextID           string `json:"next_id"`
						NextChapterTitle string `json:"next_chapter_title"`
						NextArticleTitle string `json:"next_article_title"`
					} `json:"relation"`
				} `json:"article"`
				Chapter struct {
					SourceID         int    `json:"source_id"`
					Title            string `json:"title"`
					SKU              string `json:"sku"`
					Score            string `json:"score"`
					PChapterSourceID string `json:"pchapter_source_id"`
					PChapterTitle    string `json:"p_chapter_title"`
					ChapterStatus    int    `json:"chapter_status"`
				} `json:"chapter"`
				Audio struct {
					URL         string `json:"url"`
					DownloadURL string `json:"download_url"`
					Size        int    `json:"size"`
					Title       string `json:"title"`
					Time        string `json:"time"`
					MD5         string `json:"md5"`
					Dubber      string `json:"dubber"`
					ID          string `json:"id"`
					Status      int    `json:"status"`
				} `json:"audio"`
				Video struct {
					ID    string `json:"id"`
					MD5   string `json:"md5"`
					URL   string `json:"url"`
					Cover struct {
						Type int    `json:"type"`
						ID   int    `json:"id"`
						URL  string `json:"url"`
					} `json:"cover"`
					Width     int    `json:"width"`
					Height    int    `json:"height"`
					Size      int    `json:"size"`
					Time      string `json:"time"`
					HlsMedias []struct {
						Quality string `json:"quality"`
						Size    int    `json:"size"`
						URL     string `json:"url"`
					} `json:"hls_medias"`
					HlsVid       string      `json:"hls_vid"`
					Version      int         `json:"version"`
					Medias       interface{} `json:"medias"`
					MediaOpen    string      `json:"media_open"`
					CouldPreview int         `json:"could_preview"`
					Preview      struct {
						Duration int `json:"duration"`
						Medias   []struct {
							Quality string `json:"quality"`
							Size    int    `json:"size"`
							URL     string `json:"url"`
						} `json:"medias"`
					} `json:"preview"`
					Subtitles struct {
						Rights  interface{}   `json:"rights"`
						Preview []interface{} `json:"preview"`
					} `json:"subtitles"`
					Status int `json:"status"`
				} `json:"video"`
				Files []interface{} `json:"files"`
				Extra struct {
					Process struct {
						ArticleID     string `json:"article_id"`
						LearnPercent  int    `json:"learn_percent"`
						ArticleOffset struct {
							CurOffset   int `json:"cur_offset"`
							MaxOffset   int `json:"max_offset"`
							Length      int `json:"length"`
							Version     int `json:"version"`
							Process     int `json:"process"`
							LearnTime   int `json:"learn_time"`
							LearnStatus int `json:"learn_status"`
						} `json:"article_offset"`
						AudioOffset struct {
							CurOffset   int `json:"cur_offset"`
							MaxOffset   int `json:"max_offset"`
							Length      int `json:"length"`
							Version     int `json:"version"`
							Process     int `json:"process"`
							LearnTime   int `json:"learn_time"`
							LearnStatus int `json:"learn_status"`
						} `json:"audio_offset"`
						VideoOffset struct {
							CurOffset   int `json:"cur_offset"`
							MaxOffset   int `json:"max_offset"`
							Length      int `json:"length"`
							Version     int `json:"version"`
							Process     int `json:"process"`
							LearnTime   int `json:"learn_time"`
							LearnStatus int `json:"learn_status"`
						} `json:"video_offset"`
					} `json:"process"`
					IsLast bool `json:"is_last"`
					Fav    struct {
						HasDone    bool `json:"has_done"`
						TotalCount int  `json:"total_count"`
						FavID      int  `json:"fav_id"`
						FavType    int  `json:"fav_type"`
					} `json:"fav"`
					IsShow      bool          `json:"IsShow"`
					Attachments []interface{} `json:"attachments"`
				} `json:"extra"`
				AnyreadTotal int  `json:"anyread_total"`
				AnyreadUsed  int  `json:"anyread_used"`
				AnyreadHit   bool `json:"anyread_hit"`
			} `json:"article_list"`
		} `json:"list"`
		HasChapter   bool `json:"has_chapter"`
		IsShow       bool `json:"is_show"`
		AnyreadTotal int  `json:"anyread_total"`
		AnyreadUsed  int  `json:"anyread_used"`

		Extra struct {
			Cost      float64 `json:"cost"`
			RequestID string  `json:"request-id"`
		} `json:"extra"`
	} `json:"data"`
}
