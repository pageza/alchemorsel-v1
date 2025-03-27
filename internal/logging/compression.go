package logging

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// LogCompressor handles log file compression
type LogCompressor struct {
	logDir string
}

// NewLogCompressor creates a new log compressor
func NewLogCompressor(logDir string) *LogCompressor {
	return &LogCompressor{
		logDir: logDir,
	}
}

// CompressOldLogs compresses log files older than 7 days
func (lc *LogCompressor) CompressOldLogs() error {
	files, err := filepath.Glob(filepath.Join(lc.logDir, "*.log"))
	if err != nil {
		return fmt.Errorf("failed to find log files: %w", err)
	}

	cutoff := time.Now().AddDate(0, 0, -7)

	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			continue
		}

		if info.ModTime().Before(cutoff) {
			if err := lc.compressFile(file); err != nil {
				return err
			}
		}
	}

	return nil
}

// compressFile compresses a single log file
func (lc *LogCompressor) compressFile(filePath string) error {
	// Open the source file
	source, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer source.Close()

	// Create the compressed file
	compressedPath := filePath + ".gz"
	compressed, err := os.Create(compressedPath)
	if err != nil {
		return fmt.Errorf("failed to create compressed file: %w", err)
	}
	defer compressed.Close()

	// Create gzip writer
	gzipWriter := gzip.NewWriter(compressed)
	defer gzipWriter.Close()

	// Copy the contents
	if _, err := io.Copy(gzipWriter, source); err != nil {
		return fmt.Errorf("failed to compress file: %w", err)
	}

	// Remove the original file
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to remove original file: %w", err)
	}

	return nil
}

// DecompressFile decompresses a compressed log file
func (lc *LogCompressor) DecompressFile(compressedPath string) error {
	// Open the compressed file
	compressed, err := os.Open(compressedPath)
	if err != nil {
		return fmt.Errorf("failed to open compressed file: %w", err)
	}
	defer compressed.Close()

	// Create gzip reader
	gzipReader, err := gzip.NewReader(compressed)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzipReader.Close()

	// Create the decompressed file
	decompressedPath := compressedPath[:len(compressedPath)-3]
	decompressed, err := os.Create(decompressedPath)
	if err != nil {
		return fmt.Errorf("failed to create decompressed file: %w", err)
	}
	defer decompressed.Close()

	// Copy the contents
	if _, err := io.Copy(decompressed, gzipReader); err != nil {
		return fmt.Errorf("failed to decompress file: %w", err)
	}

	return nil
}
