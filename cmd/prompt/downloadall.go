package prompt

import (
	"fmt"
	"os"

	"github.com/manifoldco/promptui"
)

// SelectDownLoadAllOrSelectArticles show select promt to choose what to do on selected column
func SelectDownLoadAllOrSelectArticles(title string) int {

	var options = []struct {
		Text  string
	}{
		{"返回上一级"},
		{"下载当前专栏所有文章"},
		{"选择文章"},
	}

	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "\U00002714 {{ .Text | red }}",
		Inactive: "{{if eq .Value 0}} {{ .Text | green }} {{else}} {{ .Text | cyan }} {{end}}",
	}

	prompt := promptui.Select{
		Label:        fmt.Sprintf("当前选中的专栏为: %s, 请继续选择：", title),
		Items:        options,
		Templates:    templates,
		Size:         len(options),
		HideSelected: true,
	}

	index, _, err := prompt.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		os.Exit(1)
	}
	return index
}
