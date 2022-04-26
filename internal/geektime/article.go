package geektime

import (
	"github.com/go-resty/resty/v2"
)

const (
	// ArticlesPath ...
	ArticlesPath = "/serv/v1/column/articles"
	// ArticleInfoPath ...
	ArticleInfoPath = "/serv/v3/article/info"
)

// ArticleSummary ...
type ArticleSummary struct {
	AID   int
	Title string
}

// VideoInfo ...
type VideoInfo struct {
	M3U8URL string
	Size    int
}

// GetArticles call geektime api to get article list
func GetArticles(cid string, client *resty.Client) ([]ArticleSummary, error) {
	ok, err := Auth(client.Cookies)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrAuthFailed
	}

	var result struct {
		Code int `json:"code"`
		Data struct {
			List []struct {
				ID           int    `json:"id"`
				ArticleTitle string `json:"article_title"`
			} `json:"list"`
		} `json:"data"`
	}
	_, err = client.R().
		SetBody(
			map[string]interface{}{
				"cid":    cid,
				"order":  "earliest",
				"prev":   0,
				"sample": false,
				"size":   500,
			}).
		SetResult(&result).
		Post(ArticlesPath)

	if err != nil {
		return nil, err
	}

	if result.Code == 0 {
		var articles []ArticleSummary
		for _, v := range result.Data.List {
			articles = append(articles, ArticleSummary{
				AID:   v.ID,
				Title: v.ArticleTitle,
			})
		}
		return articles, nil
	}

	return nil, ErrGeekTimeAPIBadCode{ArticlesPath, result.Code, ""}
}

// GetVideoInfo call geektime api to get video info
func GetVideoInfo(aid int, quality string, client *resty.Client) (VideoInfo, error) {
	var videoInfo VideoInfo
	ok, err := Auth(client.Cookies)
	if err != nil {
		return videoInfo, err
	}
	if !ok {
		return videoInfo, ErrAuthFailed
	}

	var result struct {
		Code int `json:"code"`
		Data struct {
			Info struct {
				ID    int    `json:"id"`
				Title string `json:"title"`
				Video struct {
					HLSVideos []struct {
						Size    int    `json:"size"`
						Quality string `json:"quality"`
						URL     string `json:"url"`
					} `json:"hls_medias"`
				} `json:"video"`
			} `json:"info"`
		} `json:"data"`
	}
	_, err = client.R().
		SetBody(
			map[string]interface{}{
				"id": aid,
			}).
		SetResult(&result).
		Post(ArticleInfoPath)

	if err != nil {
		return videoInfo, err
	}

	if result.Code == 0 {
		for _, v := range result.Data.Info.Video.HLSVideos {
			if v.Quality == quality {
				return VideoInfo{
					v.URL,
					v.Size,
				}, nil
			}
		}
	}

	return videoInfo, ErrGeekTimeAPIBadCode{ArticleInfoPath, result.Code, ""}
}
