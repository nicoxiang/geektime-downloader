package ui

import (
	"fmt"

	"github.com/manifoldco/promptui"
	"github.com/nicoxiang/geektime-downloader/internal/geektime"
)

type articleOpsOption struct {
	Text  string
	Value int
}

func ProductAction(product geektime.Course) (int, error) {
	options := make([]articleOpsOption, 3)
	options[0] = articleOpsOption{"重新选择课程", 0}
	if geektime.IsTextCourse(product) {
		options[1] = articleOpsOption{"下载当前专栏所有文章", 1}
		options[2] = articleOpsOption{"选择文章", 2}
	} else {
		options[1] = articleOpsOption{"下载所有视频", 1}
		options[2] = articleOpsOption{"选择视频", 2}
	}
	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "{{ `>` | red }} {{ .Text | red }}",
		Inactive: "{{if eq .Value 0}} {{ .Text | green }} {{else}} {{ .Text }} {{end}}",
	}
	prompt := promptui.Select{
		Label:        fmt.Sprintf("当前选中的专栏为: %s, 请继续选择：", product.Title),
		Items:        options,
		Templates:    templates,
		Size:         len(options),
		HideSelected: true,
		Stdout:       NoBellStdout,
	}
	index, _, err := prompt.Run()
	if err != nil {
		return 0, err
	}
	return index, nil
}
