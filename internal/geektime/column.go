package geektime

import (
	"errors"

	"github.com/go-resty/resty/v2"
)

// ColumnSummary Mini column struct
type ColumnSummary struct {
	CID        int
	Title      string
	AuthorName string
	Articles   []ArticleSummary
}

// GetColumnList call geektime api to get column list
func GetColumnList(client *resty.Client) ([]ColumnSummary, error) {
	var result struct {
		Code int `json:"code"`
		Data struct {
			Products []struct {
				ID     int    `json:"id"`
				Title  string `json:"title"`
				Author struct {
					Name string `json:"name"`
				} `json:"author"`
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
				"type":             "c1",
				"with_learn_count": 1,
			}).
		SetResult(&result).
		Post("/serv/v3/learn/product")

	if err != nil {
		return nil, err
	}

	if result.Code == 0 {
		var products []ColumnSummary
		for _, v := range result.Data.Products {
			products = append(products, ColumnSummary{
				CID:        v.ID,
				Title:      v.Title,
				AuthorName: v.Author.Name,
			})
		}
		return products, nil
	}
	return nil, errors.New("Call geektime product api failed")
}
