package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

const (
	maxSizeBytes = 10 * 1024 * 1024 // 10 MB in bytes
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: mov_to_mp4 <input.mov>")
		os.Exit(1)
	}

	inputFile := os.Args[1]
	outputFile := changeExtension(inputFile, ".mp4")

	if !checkFFmpeg() {
		installFFmpeg()
	}

	// First try with original resolution
	convertVideo(inputFile, outputFile, "")

	// If file is too large, then try with 80% resolution
    scalingFactors := []float64{0.8, 0.6, 0.4, 0.2}
	for _, factor := range scalingFactors {
		if fileSize(outputFile) <= maxSizeBytes {
			break // File is under the size limit, we're done
		}

		percentage := int(factor * 100)
		fmt.Printf("File exceeds %dMB, retrying with %d%% resolution...\n", maxSizeBytes/(1024*1024), percentage)
		scaleFilter := fmt.Sprintf("scale=trunc(iw*%.1f/2)*2:trunc(ih*%.1f/2)*2", factor, factor)
		convertVideo(inputFile, outputFile, scaleFilter)
	}

	// If the file is still too large after trying all resolutions
    if fileSize(outputFile) > maxSizeBytes {
        fmt.Printf("Warning: File still exceeds %dMB after trying all resolution reductions.\n", maxSizeBytes/(1024*1024))
    } else {
        fileSizeKB := float64(fileSize(outputFile)) / 1024.0
        fmt.Printf("Successfully created MP4 under %dMB size limit.\n", maxSizeBytes/(1024*1024))
        fmt.Printf("Final file size: %.2f KB (%.2f MB)\n", fileSizeKB, fileSizeKB/1024.0)
    }
}

func checkFFmpeg() bool {
	cmd := exec.Command("ffmpeg", "-version")
	err := cmd.Run()
	return err == nil
}

func installFFmpeg() {
	fmt.Println("ffmpeg not found, installing via Homebrew...")
	cmd := exec.Command("brew", "install", "ffmpeg")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Println("Failed to install ffmpeg. Please install it manually.")
		os.Exit(1)
	}
}

func getVideoDuration(input string) float64 {
	cmd := exec.Command("ffprobe", "-v", "error", "-show_entries", "format=duration", "-of", "default=noprint_wrappers=1:nokey=1", input)
	output, err := cmd.Output()
	if err != nil {
		return 0
	}

	duration, err := strconv.ParseFloat(strings.TrimSpace(string(output)), 64)
	if err != nil {
		return 0
	}
	return duration
}

func calculatePercent(currentTime string, totalDuration float64) float64 {
	// Parse the time in format HH:MM:SS.MS to seconds
	var h, m, s, ms float64
	fmt.Sscanf(currentTime, "%f:%f:%f.%f", &h, &m, &s, &ms)
	currentSeconds := h*3600 + m*60 + s + ms/100

	return (currentSeconds / totalDuration) * 100
}

func convertVideo(input, output, scaleFilter string) {
	// First get the duration of the input video
	duration := getVideoDuration(input)
	if duration <= 0 {
		fmt.Println("Couldn't determine video duration. Processing without percentage...")
	}

	// Build the ffmpeg command
	args := []string{"-i", input, "-c:v", "libx264", "-preset", "medium", "-crf", "23", "-c:a", "aac", "-b:a", "128k"}

	// Add scale filter only if specified
	if scaleFilter != "" {
		args = append(args, "-vf", scaleFilter)
	}

	args = append(args, "-y", output)
	cmd := exec.Command("ffmpeg", args...)

	stderr, err := cmd.StderrPipe()
	if err != nil {
		fmt.Println("Error creating stderr pipe:", err)
		os.Exit(1)
	}

	if err := cmd.Start(); err != nil {
		fmt.Println("Error starting ffmpeg:", err)
		os.Exit(1)
	}

	progressRegex := regexp.MustCompile(`frame=.* time=(\d+:\d+:\d+.\d+)`) // Extracts time progress
	buf := make([]byte, 1024)
	for {
		n, err := stderr.Read(buf)
		if n > 0 {
			matches := progressRegex.FindStringSubmatch(string(buf[:n]))
			if len(matches) > 1 {
				currentTime := matches[1]
				if duration > 0 {
					percent := calculatePercent(currentTime, duration)
					if percent > 100 {
						percent = 100
					}

					// Determine which resolution we're using
					resolutionInfo := "original"
					if scaleFilter != "" {
						// Extract the scaling factor from the filter
						scaleFactor := 1.0
						if strings.Contains(scaleFilter, "iw*0.8") {
							scaleFactor = 0.8
						} else if strings.Contains(scaleFilter, "iw*0.6") {
							scaleFactor = 0.6
						} else if strings.Contains(scaleFilter, "iw*0.4") {
							scaleFactor = 0.4
						} else if strings.Contains(scaleFilter, "iw*0.2") {
							scaleFactor = 0.2
						}
						resolutionInfo = fmt.Sprintf("%d%%", int(scaleFactor*100))
					}

					fmt.Printf("\rProcessing: %s (%.1f%%) [%s resolution]", currentTime, percent, resolutionInfo)
				} else {
					fmt.Printf("\rProcessing: %s", currentTime)
				}
			}
		}
		if err != nil {
			break
		}
	}

	if err := cmd.Wait(); err != nil {
		fmt.Println("\nConversion failed:", err)
		os.Exit(1)
	}
	fmt.Println("\nConversion complete!")
}


func fileSize(filename string) int64 {
	info, err := os.Stat(filename)
	if err != nil {
		return 0
	}
	return info.Size()
}

func changeExtension(file, newExt string) string {
    // Get the directory and filename separately
    dir := filepath.Dir(file)
    fileName := filepath.Base(file)

    // Replace extension while preserving the directory
    newFileName := fileName[:len(fileName)-len(filepath.Ext(fileName))] + newExt

    // Join the directory with the new filename
    return filepath.Join(dir, newFileName)
}