package geektime

import (
	"strconv"

	"github.com/go-resty/resty/v2"
	"github.com/nicoxiang/geektime-downloader/internal/geektime/response"
)

const (
	// DefaultBaseURL ...
	DefaultBaseURL = "https://time.geekbang.org"

	// V1ColumnArticlesPath get all articles summary info in one column
	V1ColumnArticlesPath = "/serv/v1/column/articles"
	// V1ArticlePath used in normal column
	V1ArticlePath = "/serv/v1/article"
	// V3ColumnInfoPath used in get normal column/video info
	V3ColumnInfoPath = "/serv/v3/column/info"
	// V3ProductInfoPath used in get daily lesson, qconplus product info
	V3ProductInfoPath = "/serv/v3/product/info"
	// V3ArticleInfoPath used in normal video, daily lesson, qconplus
	V3ArticleInfoPath = "/serv/v3/article/info"
	// V3VideoPlayAuthPath used in normal video, daily lesson, qconplus video play auth
	V3VideoPlayAuthPath = "/serv/v3/source_auth/video_play_auth"

	// GeekBangCookieDomain ...
	GeekBangCookieDomain = ".geekbang.org"

	// GCID ...
	GCID = "GCID"
	// GCESS ...
	GCESS = "GCESS"
)

// Course ...
type Course struct {
	Access   bool
	ID       int
	Title    string
	Type     string
	IsVideo  bool
	Articles []Article
}

// Article ...
type Article struct {
	AID          int
	SectionTitle string
	Title        string
}

// CourseInfo get narmal geektime course info
func (c *Client) CourseInfo(productID int) (Course, error) {
	var p Course
	var err error
	p, err = c.columnInfo(productID)
	if err != nil {
		return p, err
	}

	var articles []Article
	articles, err = c.columnArticles(productID)
	if err != nil {
		return p, err
	}
	p.Articles = articles

	return p, nil
}

// V1ArticleInfo ...
func (c *Client) V1ArticleInfo(articleID int) (response.V1ArticleResponse, error) {
	var res response.V1ArticleResponse
	r := c.newRequest(
		resty.MethodPost,
		DefaultBaseURL,
		V1ArticlePath,
		nil,
		map[string]interface{}{
			"id":                strconv.Itoa(articleID),
			"include_neighbors": true,
			"is_freelyread":     true,
			"reverse":           false,
		},
		&res,
	)
	if _, err := do(r); err != nil {
		return response.V1ArticleResponse{}, err
	}
	return res, nil
}

// ProductInfo used to get daily lesson or qconplus product info
func (c *Client) ProductInfo(productID int) (response.V3ProductInfoResponse, error) {
	var res response.V3ProductInfoResponse
	r := c.newRequest(
		resty.MethodPost,
		DefaultBaseURL,
		V3ProductInfoPath,
		nil,
		map[string]interface{}{
			"id": productID,
		},
		&res,
	)
	if _, err := do(r); err != nil {
		return response.V3ProductInfoResponse{}, err
	}
	return res, nil
}

// V3ArticleInfo used to get daily lesson or qconplus article info
func (c *Client) V3ArticleInfo(articleID int) (response.V3ArticleInfoResponse, error) {
	var res response.V3ArticleInfoResponse
	r := c.newRequest(
		resty.MethodPost,
		DefaultBaseURL,
		V3ArticleInfoPath,
		nil,
		map[string]interface{}{
			"id": articleID,
		},
		&res,
	)
	if _, err := do(r); err != nil {
		return response.V3ArticleInfoResponse{}, err
	}
	return res, nil
}

// VideoPlayAuth get play auth string
func (c *Client) VideoPlayAuth(articleID, sourceType int, videoID string) (string, error) {
	var res response.V3VideoPlayAuthResponse
	r := c.newRequest(
		resty.MethodPost,
		DefaultBaseURL,
		V3VideoPlayAuthPath,
		nil,
		map[string]interface{}{
			"aid":         articleID,
			"source_type": sourceType,
			"video_id":    videoID,
		},
		&res,
	)
	if _, err := do(r); err != nil {
		return "", err
	}
	return res.Data.PlayAuth, nil
}

// columnInfo get normal column info, like v3 product info
func (c *Client) columnInfo(productID int) (Course, error) {
	var res response.V3ColumnInfoResponse
	r := c.newRequest(
		resty.MethodPost,
		DefaultBaseURL,
		V3ColumnInfoPath,
		nil,
		map[string]interface{}{
			"product_id":             productID,
			"with_recommend_article": true,
		},
		&res,
	)
	if _, err := do(r); err != nil {
		return Course{}, err
	}

	return Course{
		Access:  res.Data.Extra.Sub.AccessMask > 0,
		ID:      res.Data.ID,
		Type:    res.Data.Type,
		Title:   res.Data.Title,
		IsVideo: res.Data.IsVideo,
	}, nil
}

// columnArticles call geektime api to get article list
func (c *Client) columnArticles(cid int) ([]Article, error) {
	res := &response.V1ColumnArticlesResponse{}
	r := c.newRequest(
		resty.MethodPost,
		DefaultBaseURL,
		V1ColumnArticlesPath,
		nil,
		map[string]interface{}{
			"cid":    strconv.Itoa(cid),
			"order":  "earliest",
			"prev":   0,
			"sample": false,
			"size":   500, //get all articles
		},
		res,
	)
	if _, err := do(r); err != nil {
		return nil, err
	}

	var articles []Article
	for _, v := range res.Data.List {
		articles = append(articles, Article{
			AID:   v.ID,
			Title: v.ArticleTitle,
		})
	}
	return articles, nil
}
