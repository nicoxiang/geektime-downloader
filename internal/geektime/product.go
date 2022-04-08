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

type ProductRequest struct {
	Code int `json:"code"`
	Data struct {
		List []struct {
			Score int `json:"score"`
		} `json:"list"`
		Products []struct {
			ID     int    `json:"id"`
			Title  string `json:"title"`
			Author struct {
				Name string `json:"name"`
			} `json:"author"`
			Type string `json:"type"`
		} `json:"products"`
		Page struct {
			More bool `json:"more"`
		} `json:"page"`
	} `json:"data"`
	Error struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	} `json:"error"`
}

// GetProductList call geektime api to get product list
func GetProductList(client *resty.Client) ([]Product, error) {
	if !Auth(client.Cookies) {
		return nil, ErrAuthFailed
	}
	var products []Product = make([]Product, 0)
	addProducts(client, 0, &products)
	return products, nil
}

func addProducts(client *resty.Client, prev int, products *[]Product) {
	var result ProductRequest
	_, err := client.R().
		SetBody(
			map[string]interface{}{
				"desc":             false,
				"expire":           1,
				"last_learn":       0,
				"learn_status":     0,
				"prev":             prev,
				"size":             20,
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
		for _, v := range result.Data.Products {
			// For now we can only download column and video
			if v.Type == "c1" || v.Type == "c3" {
				*products = append(*products, Product{
					ID:         v.ID,
					Title:      v.Title,
					AuthorName: v.Author.Name,
					Type:       v.Type,
				})
			}
		}
		if result.Data.Page.More {
			score := result.Data.List[0].Score
			addProducts(client, score, products)
		}
	} else {
		panic("make geektime product api call failed")
	}
}
