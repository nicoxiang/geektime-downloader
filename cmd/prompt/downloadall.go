package prompt

import (
	"fmt"
	"os"

	"github.com/manifoldco/promptui"
)

type selectOption struct {
	Value int
	Text  string
}

var downLoadAllOrSelectArticlesOptions = []selectOption{
	{
		Value: 0,
		Text:  "返回上一级",
	},
	{
		Value: 1,
		Text:  "下载当前专栏所有文章",
	},
	{
		Value: 2,
		Text:  "选择文章",
	},
}

// Show select promt to choose what to do on selected column
func PromptSelectDownLoadAllOrSelectArticles(title string) int {
	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "\U00002714 {{ .Text | red }}",
		Inactive: "{{if eq .Value 0}} {{ .Text | green }} {{else}} {{ .Text | cyan }} {{end}}",
	}

	prompt := promptui.Select{
		Label:        fmt.Sprintf("当前选中的专栏为: %s, 请继续选择：", title),
		Items:        downLoadAllOrSelectArticlesOptions,
		Templates:    templates,
		Size:         len(downLoadAllOrSelectArticlesOptions),
		HideSelected: true,
	}

	index, _, err := prompt.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		os.Exit(1)
	}
	return index
}
