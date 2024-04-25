package geektime

import (
	"strconv"

	"github.com/go-resty/resty/v2"
	"github.com/nicoxiang/geektime-downloader/internal/geektime/response"
)

const (
	// GeekBangEnterpriseBaseURL is geekbang enterprise base URL
	GeekBangEnterpriseBaseURL = "https://b.geekbang.org"
	// V1EnterpriseCourseInfoPath used in enterprise course product info
	V1EnterpriseCourseInfoPath = "/app/v1/course/info"
	// V1EnterpriseArticlesInfoPath used in enterprise course articles info
	V1EnterpriseArticlesInfoPath = "/app/v1/course/articles"
	// V1EnterpriseArticleDetailPath used in enterprise course article detail info
	V1EnterpriseArticleDetailPath = "/app/v1/article/detail"
	// V1EnterpriseVideoPlayAuthPath used in enterprise course video play auth
	V1EnterpriseVideoPlayAuthPath = "/app/v1/source_auth/video_play_auth"
)

// EnterpriseCourseInfo get enterprise course info
func (c *Client) EnterpriseCourseInfo(id int) (Course, error) {
	var p Course
	var err error
	p, err = c.enterpriseCourseInfo(id)
	if err != nil {
		return p, err
	}

	var articles []Article
	articles, err = c.enterpriseCourseArticles(id)
	if err != nil {
		return p, err
	}
	p.Articles = articles

	return p, nil
}

// V1EnterpriseArticleDetail get enterprise article detail
func (c *Client) V1EnterpriseArticleDetail(articleID string) (response.V1EnterpriseArticlesDetailResponse, error) {
	var res response.V1EnterpriseArticlesDetailResponse
	r := c.newRequest(
		resty.MethodPost,
		GeekBangEnterpriseBaseURL,
		V1EnterpriseArticleDetailPath,
		nil,
		map[string]interface{}{
			"article_id": articleID,
		},
		&res,
	)
	if _, err := do(r); err != nil {
		return response.V1EnterpriseArticlesDetailResponse{}, err
	}
	return res, nil
}

// EnterpriseVideoPlayAuth get enterprise play auth string
func (c *Client) EnterpriseVideoPlayAuth(articleID, videoID string) (string, error) {
	var res response.V3VideoPlayAuthResponse
	r := c.newRequest(
		resty.MethodPost,
		GeekBangEnterpriseBaseURL,
		V1EnterpriseVideoPlayAuthPath,
		nil,
		map[string]interface{}{
			"aid":      articleID,
			"video_id": videoID,
		},
		&res,
	)
	if _, err := do(r); err != nil {
		return "", err
	}
	return res.Data.PlayAuth, nil
}

func (c *Client) enterpriseCourseInfo(productID int) (Course, error) {
	var res response.V1EnterpriseProductInfoResponse

	r := c.newRequest(
		resty.MethodPost,
		GeekBangEnterpriseBaseURL,
		V1EnterpriseCourseInfoPath,
		nil,
		map[string]interface{}{
			"id": productID,
		},
		&res,
	)

	if _, err := do(r); err != nil {
		return Course{}, err
	}

	return Course{
		Access:  res.Data.Extra.IsMyCourse,
		ID:      productID,
		Title:   res.Data.Title,
		Type:    "",
		IsVideo: true,
	}, nil
}

func (c *Client) enterpriseCourseArticles(productID int) ([]Article, error) {
	var res response.V1EnterpriseArticlesResponse
	r := c.newRequest(
		resty.MethodPost,
		GeekBangEnterpriseBaseURL,
		V1EnterpriseArticlesInfoPath,
		nil,
		map[string]interface{}{
			"id": productID,
		},
		&res,
	)

	if _, err := do(r); err != nil {
		return nil, err
	}

	var articles []Article

	for _, sections := range res.Data.List {
		for _, a := range sections.ArticleList {
			articleID, _ := strconv.Atoi(a.Article.ID)
			articles = append(articles, Article{
				AID:          articleID,
				SectionTitle: sections.Title,
				Title:        a.Article.Title,
			})
		}
	}
	return articles, nil
}
