package prompt

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/manifoldco/promptui"
)

// Show prompt to let user input password
func PromptGetPwd() string {
	validate := func(input string) error {
		if strings.TrimSpace(input) == "" {
			return errors.New("密码不能为空")
		}
		return nil
	}

	prompt := promptui.Prompt{
		Label:    "请输入密码",
		Validate: validate,
		Mask:     '*',
	}

	result, err := prompt.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		os.Exit(1)
	}

	return result
}