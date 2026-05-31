package ffmpeg

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/nicoxiang/geektime-downloader/internal/pkg/logger"
)

// MP4Extension mp4 file extension
const MP4Extension = ".mp4"

// ConvertToMP4 convert ts file to mp4 using ffmpeg
// Input: source file path (ts), output directory
// Output: mp4 file path, error
func ConvertToMP4(tsFilePath, outputDir string) (string, error) {
	// Validate input file exists
	if _, err := os.Stat(tsFilePath); os.IsNotExist(err) {
		return "", fmt.Errorf("source ts file not found: %s", tsFilePath)
	}

	// Generate output mp4 file path
	baseName := filepath.Base(tsFilePath)
	mp4FileName := baseName[:len(baseName)-len(".ts")] + MP4Extension
	mp4FilePath := filepath.Join(outputDir, mp4FileName)

	// Check if mp4 already exists
	if _, err := os.Stat(mp4FilePath); err == nil {
		logger.Infof("MP4 file already exists, skipping conversion: %s", mp4FilePath)
		return mp4FilePath, nil
	}

	// Find ffmpeg executable
	ffmpegPath, err := findFFmpeg()
	if err != nil {
		return "", fmt.Errorf("ffmpeg not found: %w", err)
	}

	logger.Infof("Converting %s to %s", baseName, mp4FileName)

	// Prepare ffmpeg command
	// -i: input file
	// -c copy: copy streams without re-encoding (fast)
	// -y: overwrite output file without asking
	args := []string{
		"-i", tsFilePath,
		"-c", "copy",
		"-y",
		mp4FilePath,
	}

	cmd := exec.Command(ffmpegPath, args...)

	// Capture output for logging
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("ffmpeg conversion failed: %w, output: %s", err, string(output))
	}

	logger.Infof("Successfully converted to: %s", mp4FilePath)

	return mp4FilePath, nil
}

// findFFmpeg finds the ffmpeg executable in the system
func findFFmpeg() (string, error) {
	// Try common ffmpeg executable names
	names := []string{"ffmpeg", "ffmpeg.exe"}

	// First, check if ffmpeg is in PATH
	for _, name := range names {
		if path, err := exec.LookPath(name); err == nil {
			return path, nil
		}
	}

	// For Windows, try to find in common installation paths
	if runtime.GOOS == "windows" {
		commonPaths := []string{
			`C:\Program Files\ffmpeg\bin\ffmpeg.exe`,
			`C:\ffmpeg\bin\ffmpeg.exe`,
			filepath.Join(os.Getenv("LOCALAPPDATA"), "ffmpeg", "bin", "ffmpeg.exe"),
		}
		for _, path := range commonPaths {
			if _, err := os.Stat(path); err == nil {
				return path, nil
			}
		}
	}

	return "", fmt.Errorf("ffmpeg executable not found in PATH or common locations. Please install ffmpeg from https://ffmpeg.org/download.html")
}

// IsFFmpegAvailable checks if ffmpeg is available in the system
func IsFFmpegAvailable() bool {
	_, err := findFFmpeg()
	return err == nil
}
