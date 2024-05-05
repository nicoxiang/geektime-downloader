package geektime

import (
	"github.com/go-resty/resty/v2"
	"github.com/nicoxiang/geektime-downloader/internal/geektime/response"
)

const (
	// GeekBangUniversityBaseURL ...
	GeekBangUniversityBaseURL = "https://u.geekbang.org"

	// UniversityV1VideoPlayAuthPath used in university video play auth
	UniversityV1VideoPlayAuthPath = "/serv/v1/video/play-auth"
	// UniversityV1MyClassInfoPath get university class info and all articles info in it
	UniversityV1MyClassInfoPath = "/serv/v1/myclass/info"
)

// UniversityCourseInfo get university class info
func (c *Client) UniversityCourseInfo(classID int) (Course, error) {
	var p Course

	var res response.V1MyClassInfoResponse
	r := c.newRequest(
		resty.MethodPost,
		GeekBangUniversityBaseURL,
		UniversityV1MyClassInfoPath,
		nil,
		map[string]interface{}{
			"class_id": classID,
		},
		&res,
	)

	resp, err := do(r)
	if err != nil {
		return p, err
	}

	if res.Code != 0 {
		if res.Error.Code == -5001 {
			p.Access = false
			return p, nil
		}
		return p, ErrGeekTimeAPIBadCode{UniversityV1MyClassInfoPath, resp.String()}
	}

	p = Course{
		Access:  true,
		ID:      classID,
		Title:   res.Data.Title,
		Type:    "",
		IsVideo: true,
	}
	var articles []Article
	for _, lesson := range res.Data.Lessons {
		for _, article := range lesson.Articles {
			// ONLY download university video lessons
			if article.VideoTime > 0 {
				articles = append(articles, Article{
					AID:   article.ArticleID,
					Title: article.ArticleTitle,
				})
			}
		}
	}
	p.Articles = articles

	return p, nil
}

// UniversityVideoPlayAuth get university play auth string
func (c *Client) UniversityVideoPlayAuth(articleID, classID int) (response.V1VideoPlayAuthResponse, error) {
	var res response.V1VideoPlayAuthResponse
	r := c.newRequest(
		resty.MethodPost,
		GeekBangUniversityBaseURL,
		UniversityV1VideoPlayAuthPath,
		nil,
		map[string]interface{}{
			"article_id": articleID,
			"class_id":   classID,
		},
		&res,
	)
	if _, err := do(r); err != nil {
		return response.V1VideoPlayAuthResponse{}, err
	}
	return res, nil
}
