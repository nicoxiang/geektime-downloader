package progressbar

import (
	"time"

	"github.com/cheggaaa/pb/v3"
)

// New provides a custom progressbar to measure byte
// throughput with recommended defaults.
func New(size int64, prefix string) *pb.ProgressBar {
	bar := pb.New64(size)
	bar.SetRefreshRate(time.Second)
	bar.Set(pb.Bytes, true)
	bar.Set(pb.SIBytesPrefix, true)
	bar.SetTemplate(pb.Simple)
	bar.Set("prefix", prefix)
	return bar
}
