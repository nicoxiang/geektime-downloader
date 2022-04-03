package prompt

import (
	"errors"
	"os"

	"github.com/manifoldco/promptui"
	"github.com/nicoxiang/geektime-downloader/internal/geektime"
)

// SelectProduct show select promt to choose a product
func SelectProduct(products []geektime.Product) int {
	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "\U00002714 {{if eq .Type `c1`}} {{ `专栏` | red }} {{else}} {{ `视频课` | red }} {{end}} {{ .Title | red }} {{ .AuthorName | red }}",
		Inactive: "{{if eq .Type `c1`}} {{ `专栏` }} {{else}} {{ `视频课` }} {{end}} {{ .Title }} {{ .AuthorName }}",
	}

	prompt := promptui.Select{
		Label:        "请选择课程: ",
		Items:        products,
		Templates:    templates,
		Size:         len(products),
		HideSelected: true,
	}

	index, _, err := prompt.Run()

	if err != nil {
		if !errors.Is(err, promptui.ErrInterrupt) {
			panic(err)
		}
		os.Exit(1)
	}
	return index
}
