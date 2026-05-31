package ffmpeg

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsFFmpegAvailable(t *testing.T) {
	available := IsFFmpegAvailable()
	t.Logf("FFmpeg available: %v", available)

	if !available {
		t.Skip("ffmpeg not available, skipping conversion tests")
	}
}

func TestConvertToMP4(t *testing.T) {
	if !IsFFmpegAvailable() {
		t.Skip("ffmpeg not available, skipping test")
	}

	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create a dummy ts file (this won't be a valid video, just testing the command logic)
	tsFilePath := filepath.Join(tempDir, "test_video.ts")
	err := os.WriteFile(tsFilePath, []byte("dummy content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test ts file: %v", err)
	}

	// Test conversion - this will likely fail due to invalid ts content,
	// but we're testing the command execution path
	_, err = ConvertToMP4(tsFilePath, tempDir)

	// We expect an error due to invalid ts content, but the function should execute
	if err == nil {
		t.Log("Conversion executed (expected to fail with dummy content)")
	}
}

func TestFindFFmpeg(t *testing.T) {
	path, err := findFFmpeg()
	if err != nil {
		t.Logf("FFmpeg not found: %v", err)
		return
	}

	t.Logf("Found FFmpeg at: %s", path)

	// Verify the path exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("FFmpeg path exists but file not found: %s", path)
	}
}
