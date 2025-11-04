package ui

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/manifoldco/promptui"
)

func ProductIDInput(selectedProductType ProductTypeSelectOption) (int, error) {
	prompt := promptui.Prompt{
		Label: fmt.Sprintf("请输入%s的课程 ID", selectedProductType.Text),
		Validate: func(s string) error {
			if strings.TrimSpace(s) == "" {
				return errors.New("课程 ID 不能为空")
			}
			if _, err := strconv.Atoi(s); err != nil {
				return errors.New("课程 ID 格式不合法")
			}
			return nil
		},
		HideEntered: true,
	}
	s, err := prompt.Run()
	if err != nil {
		return 0, err
	}
	// ignore, because checked before
	id, _ := strconv.Atoi(s)
	return id, nil
}
