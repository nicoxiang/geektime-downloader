package ui

import (
	"github.com/manifoldco/promptui"
)

type ProductTypeSelectOption struct {
	Index              int
	Text               string
	SourceType         int
	AcceptProductTypes []string
	NeedSelectArticle  bool
	IsEnterpriseMode   bool
}

func ProductTypeSelect(isEnterprise bool) (ProductTypeSelectOption, error) {

	productTypeOptions := []ProductTypeSelectOption{}

	if isEnterprise {
		productTypeOptions = append(productTypeOptions, ProductTypeSelectOption{0, "训练营", 5, []string{"c44"}, true, true}) //custom source type, not use
	} else {
		productTypeOptions = append(productTypeOptions, ProductTypeSelectOption{0, "普通课程", 1, []string{"c1", "c3"}, true, false})
		productTypeOptions = append(productTypeOptions, ProductTypeSelectOption{1, "每日一课", 2, []string{"d"}, false, false})
		productTypeOptions = append(productTypeOptions, ProductTypeSelectOption{2, "公开课", 1, []string{"p35", "p29", "p30"}, true, false})
		productTypeOptions = append(productTypeOptions, ProductTypeSelectOption{3, "大厂案例", 4, []string{"q"}, false, false})
		productTypeOptions = append(productTypeOptions, ProductTypeSelectOption{4, "训练营", 5, []string{""}, true, false}) //custom source type, not use
		productTypeOptions = append(productTypeOptions, ProductTypeSelectOption{5, "其他", 1, []string{"x", "c6"}, true, false})
	}

	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "{{ `>` | red }} {{ .Text | red }}",
		Inactive: "{{ .Text }}",
	}
	prompt := promptui.Select{
		Label:        "请选择想要下载的产品类型",
		Items:        productTypeOptions,
		Templates:    templates,
		Size:         len(productTypeOptions),
		HideSelected: true,
		Stdout:       NoBellStdout,
	}
	index, _, err := prompt.Run()
	if err != nil {
		return ProductTypeSelectOption{}, err
	}
	return productTypeOptions[index], nil
}

// IsUniversity checks if the product type is university product type
func (p *ProductTypeSelectOption) IsUniversity() bool {
	return p.Index == 4 && !p.IsEnterpriseMode
}
