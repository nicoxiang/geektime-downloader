package geektime

import (
	"github.com/go-resty/resty/v2"
)

// Product ...
type Product struct {
	ID         int
	Title      string
	AuthorName string
	Type       string
	Articles   []ArticleSummary
}

// GetProductList call geektime api to get column list
func GetProductList(client *resty.Client) ([]Product, error) {
	if !Auth(client.Cookies) {
		return nil, ErrAuthFailed
	}
	var result struct {
		Code int `json:"code"`
		Data struct {
			Products []struct {
				ID     int    `json:"id"`
				Title  string `json:"title"`
				Author struct {
					Name string `json:"name"`
				} `json:"author"`
				Type string `json:"type"`
			} `json:"products"`
		} `json:"data"`
		Error struct {
			Code int    `json:"code"`
			Msg  string `json:"msg"`
		} `json:"error"`
	}

	client.SetHeader("Referer", "https://time.geekbang.org/dashboard/course")
	_, err := client.R().
		SetBody(
			map[string]interface{}{
				"desc":             false,
				"expire":           1,
				"last_learn":       0,
				"learn_status":     0,
				"prev":             0,
				"size":             200,
				"sort":             1,
				"type":             "",
				"with_learn_count": 1,
			}).
		SetResult(&result).
		Post("/serv/v3/learn/product")

	if err != nil {
		panic(err)
	}

	if result.Code == 0 {
		var products []Product
		for _, v := range result.Data.Products {
			// For now we can only download column and video
			if v.Type == "c1" || v.Type == "c3" {
				products = append(products, Product{
					ID:         v.ID,
					Title:      v.Title,
					AuthorName: v.Author.Name,
					Type:       v.Type,
				})
			}
		}
		return products, nil
	}
	panic("make geektime product api call failed")
}
