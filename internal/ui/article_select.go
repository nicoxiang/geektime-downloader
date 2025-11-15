package ui

import (
	"github.com/manifoldco/promptui"
	"github.com/nicoxiang/geektime-downloader/internal/geektime"
)

func ArticleSelect(articles []geektime.Article) (int, error) {
	items := []geektime.Article{
		{
			AID:   -1,
			Title: "返回上一级",
		},
	}
	items = append(items, articles...)
	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "{{ `>` | red }} {{ .Title | red }}",
		Inactive: "{{if eq .AID -1}} {{ .Title | green }} {{else}} {{ .Title }} {{end}}",
	}
	prompt := promptui.Select{
		Label:        "请选择文章: ",
		Items:        items,
		Templates:    templates,
		Size:         20,
		HideSelected: true,
		CursorPos:    0,
		Stdout:       NoBellStdout,
	}
	index, _, err := prompt.Run()
	if err != nil {
		return 0, err
	}
	return index, nil
}
