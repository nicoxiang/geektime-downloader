package geektime

// ProductType geektime product type
type ProductType string

const (
	// ProductTypeColumn 专栏
	ProductTypeColumn ProductType = "c1"
	// ProductTypeNormalVideo 视频课
	ProductTypeNormalVideo ProductType = "c3"
	// ProductTypeDailyLesson 每日一课
	ProductTypeDailyLesson ProductType = "d"
	// ProductTypeQCONPlus 大厂案例
	ProductTypeQCONPlus ProductType = "q"
	// ProductTypeUniversityVideo 训练营视频，自定义类型
	ProductTypeUniversityVideo ProductType = "u"
)
