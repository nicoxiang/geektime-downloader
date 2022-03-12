package prompt

import (
	"fmt"
	"os"

	"github.com/manifoldco/promptui"
	"github.com/nicoxiang/geektime-downloader/internal/geektime"
)

func PromptSelectColumn(columns []geektime.ColumnSummary) int {
	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "\U00002714 {{ .Title | red }} {{ .AuthorName | red }}",
		Inactive: "{{ .Title | cyan }} {{ .AuthorName | cyan }}",
	}

	prompt := promptui.Select{
		Label:        "请选择专栏: ",
		Items:        columns,
		Templates:    templates,
		Size:         len(columns),
		HideSelected: true,
	}

	index, _, err := prompt.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		os.Exit(1)
	}
	return index
}
