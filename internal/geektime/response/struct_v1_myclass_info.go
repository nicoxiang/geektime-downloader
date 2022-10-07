package response

// V1MyClassInfoResponse ...
type V1MyClassInfoResponse struct {
	Code int `json:"code"`
	Data struct {
		ClassType int    `json:"class_type"`
		Title     string `json:"title"`
		Lessons   []struct {
			ChapterName string `json:"chapter_name"`
			BeginTime   int    `json:"begin_time"`
			ChapterID   int    `json:"chapter_id"`
			IndexNo     int    `json:"index_no"`
			Articles    []struct {
				ArticleID    int    `json:"article_id"`
				ArticleTitle string `json:"article_title"`
				IndexNo      int    `json:"index_no"`
				IsRead       bool   `json:"is_read"`
				IsFinish     bool   `json:"is_finish"`
				// HasNotes         bool          `json:"has_notes"`
				// IsRequired       int           `json:"is_required"`
				VideoTime int `json:"video_time"`
				// LearnTime        int           `json:"learn_time"`
				// LearnStatus      int           `json:"learn_status"`
				// MaxOffset        int           `json:"max_offset"`
				// ArticleMaxOffset int           `json:"article_max_offset"`
				// VideoMaxOffset   int           `json:"video_max_offset"`
				// ArticleLen       int           `json:"article_len"`
				// VideoLen         int           `json:"video_len"`
				// Ctime            int           `json:"ctime"`
				// Exercises        []interface{} `json:"exercises"`
			} `json:"articles"`
		} `json:"lessons"`
	} `json:"data"`
	Error struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	} `json:"error"`
}
