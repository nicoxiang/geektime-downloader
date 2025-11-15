package fsm

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/briandowns/spinner"
	"github.com/manifoldco/promptui"

	"github.com/nicoxiang/geektime-downloader/internal/config"
	"github.com/nicoxiang/geektime-downloader/internal/course"
	"github.com/nicoxiang/geektime-downloader/internal/geektime"
	"github.com/nicoxiang/geektime-downloader/internal/pkg/logger"
	"github.com/nicoxiang/geektime-downloader/internal/ui"
)

type FSMRunner struct {
	ctx                 context.Context
	currentState        State
	config              *config.AppConfig
	selectedProductType ui.ProductTypeSelectOption
	selectedProduct     geektime.Course
	sp                  *spinner.Spinner
	geektimeClient      *geektime.Client
	courseDownloader    *course.CourseDownloader
}

// NewFSMRunner creates and initializes a new FSMRunner instance
func NewFSMRunner(ctx context.Context, cfg *config.AppConfig, geektimeClient *geektime.Client) *FSMRunner {
	sp := spinner.New(spinner.CharSets[4], 100*time.Millisecond)
	return &FSMRunner{
		ctx:              ctx,
		currentState:     StateSelectProductType,
		config:           cfg,
		sp:               sp,
		geektimeClient:   geektimeClient,
		courseDownloader: course.NewCourseDownloader(ctx, cfg, geektimeClient, sp),
	}
}

// Run executes the finite state machine loop, handling user input and state transitions.
func (r *FSMRunner) Run() error {
	for {
		var err error
		switch r.currentState {
		case StateSelectProductType:
			var selectedOption ui.ProductTypeSelectOption
			selectedOption, err = ui.ProductTypeSelect(r.config.IsEnterprise)
			if err == nil {
				r.currentState = StateInputProductID
				r.selectedProductType = selectedOption
			}
		case StateInputProductID:
			var productID int
			productID, err = ui.ProductIDInput(r.selectedProductType)
			if err == nil {
				if r.selectedProductType.NeedSelectArticle {
					err = r.handleInputProductIDIfNeedSelectArticle(productID)
				} else {
					err = r.handleInputProductIDIfDownloadDirectly(productID)
				}
			}
		case StateProductAction:
			var index int
			index, err = ui.ProductAction(r.selectedProduct)
			if err == nil {
				switch index {
				case 0:
					r.currentState = StateSelectProductType
				case 1:
					err = r.handleDownloadAll()
				case 2:
					r.currentState = StateSelectArticle
				}
			}
		case StateSelectArticle:
			var index int
			index, err = ui.ArticleSelect(r.selectedProduct.Articles)
			if err == nil {
				err = r.handleSelectArticle(index)
			}
		}

		if err != nil {
			switch {
			case errors.Is(err, context.Canceled):
				// clear line
				fmt.Print("\033[1A\033[2K")
				return nil
			case errors.Is(err, promptui.ErrInterrupt):
				// clear two lines beacause promptui print one more line if interrupt
				fmt.Print("\033[1A\033[2K\033[1A\033[2K")
				return nil
			case os.IsTimeout(err):
				logger.Errorf(err, "Request timed out")
				return fmt.Errorf("请求超时")
			default:
				logger.Errorf(err, "An error occurred")
				return err
			}
		}
	}
}

func (r *FSMRunner) handleInputProductIDIfNeedSelectArticle(productID int) error {
	// choose download all or download specified article
	r.sp.Prefix = "[ 正在加载课程信息... ]"
	r.sp.Start()
	defer r.sp.Stop()

	var course geektime.Course
	var err error
	if r.selectedProductType.IsEnterpriseMode {
		// TODO: check enterprise course type
		course, err = r.geektimeClient.EnterpriseCourseInfo(productID)
		if err != nil {
			return err
		}

	} else {
		if r.selectedProductType.IsUniversity() {
			// university don't need check product type
			// if input invalid id, access mark is 0
			course, err = r.geektimeClient.UniversityCourseInfo(productID)
			if err != nil {
				return err
			}
		} else {
			course, err = r.geektimeClient.CourseInfo(productID)
			if err == nil {
				valid := r.validateProductCode(course.Type)
				// if check product type fail, re-input product
				if !valid {
					r.currentState = StateInputProductID
					return nil
				}
			} else {
				return err
			}
		}
	}
	if !course.Access {
		fmt.Fprint(os.Stderr, "尚未购买该课程\n")
		r.currentState = StateInputProductID
		return nil
	}
	r.selectedProduct = course
	r.currentState = StateProductAction
	return nil
}

func (r *FSMRunner) handleInputProductIDIfDownloadDirectly(productID int) error {
	// when product type is daily lesson or qconplus,
	// input id means product id
	// download video directly
	productInfo, err := r.geektimeClient.ProductInfo(productID)
	if err != nil {
		return err
	}

	if productInfo.Data.Info.Extra.Sub.AccessMask == 0 {
		fmt.Fprint(os.Stderr, "尚未购买该课程\n")
		r.currentState = StateInputProductID
		return nil
	}

	if r.validateProductCode(productInfo.Data.Info.Type) {
		err = r.courseDownloader.DownloadSingleVideoProduct(productInfo.Data.Info.Title,
			productInfo.Data.Info.Article.ID,
			r.selectedProductType.SourceType)
		if err != nil {
			return err
		}
	}
	r.currentState = StateInputProductID
	return nil
}

// validateProductCode checks if the product code field in the response body returned by the API
// exists in the selected product's accepted product types list.
func (r *FSMRunner) validateProductCode(productCode string) bool {
	for _, pt := range r.selectedProductType.AcceptProductTypes {
		if pt == productCode {
			return true
		}
	}
	fmt.Fprint(os.Stderr, "\r输入的课程 ID 有误\n")
	return false
}

func (r *FSMRunner) handleSelectArticle(index int) error {
	if index == 0 {
		r.currentState = StateProductAction
		return nil
	}
	a := r.selectedProduct.Articles[index-1]

	err := r.courseDownloader.DownloadArticle(r.selectedProduct, r.selectedProductType, a, true)
	if err != nil {
		return err
	}
	fmt.Printf("\r%s 下载完成", a.Title)
	time.Sleep(time.Second)
	r.currentState = StateSelectArticle
	return nil
}

// handleDownloadAll manages the bulk download process for all articles in a selected product (course).
// Returns an error if any step in the download process fails.
func (r *FSMRunner) handleDownloadAll() error {
	if err := r.courseDownloader.DownloadAll(r.selectedProduct, r.selectedProductType); err != nil {
		return err
	}
	r.currentState = StateSelectProductType
	return nil
}
