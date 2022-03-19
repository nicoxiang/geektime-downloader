package loader

import (
	"time"

	"github.com/briandowns/spinner"
)

// New spinner without prefix
func NewSpinner() *spinner.Spinner {
	s := spinner.New(spinner.CharSets[70], 100*time.Millisecond)
	return s
}

// Wrapper method that load spinner when run inner function
func Run(s *spinner.Spinner, prefix string, inner func()) {
	s.Prefix = prefix
	s.Start()
	inner()
	s.Stop()
}
