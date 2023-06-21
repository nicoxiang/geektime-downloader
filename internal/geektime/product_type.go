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
	// ProductTypeOpenCourse 公开课(视频)
	ProductTypeOpenCoureVideo ProductType = "p35"
	// ProductTypeOpenCourse 公开课(文字)
	ProductTypeOpenCoureText ProductType = "p29"
	// ProductTypeMeetting 会议
	ProductTypeMeetting = "c6"
)
