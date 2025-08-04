package audio

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gen2brain/malgo"
)

// AudioRecorder handles audio recording functionality
type AudioRecorder struct {
	context     *malgo.AllocatedContext
	device      *malgo.Device
	isRecording bool
	outputFile  *os.File
	stopChan    chan struct{}
}

// RecordingMetadata contains information about the recording
type RecordingMetadata struct {
	CourseID     string    `json:"courseId"`
	LectureTitle string    `json:"lectureTitle"`
	Duration     int       `json:"duration"`
	Timestamp    time.Time `json:"timestamp"`
	FilePath     string    `json:"filePath"`
	SampleRate   int       `json:"sampleRate"`
	Channels     int       `json:"channels"`
	Format       string    `json:"format"`
}

// NewAudioRecorder creates a new audio recorder instance
func NewAudioRecorder() (*AudioRecorder, error) {
	context, err := malgo.InitContext(nil, malgo.ContextConfig{}, func(message string) {
		log.Printf("Audio context: %s", message)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create audio context: %w", err)
	}

	return &AudioRecorder{
		context:  context,
		stopChan: make(chan struct{}),
	}, nil
}

// GetAudioDevices returns available audio input devices
func (ar *AudioRecorder) GetAudioDevices() ([]map[string]interface{}, error) {
	devices, err := ar.context.Devices(malgo.Capture)
	if err != nil {
		return nil, fmt.Errorf("failed to get audio devices: %w", err)
	}

	var result []map[string]interface{}
	for i, device := range devices {
		deviceInfo := map[string]interface{}{
			"deviceId": fmt.Sprintf("device_%d", i),
			"label":    device.Name(),
			"kind":     "audioinput",
			"index":    i,
		}
		result = append(result, deviceInfo)
	}

	// If no devices found, return a default device
	if len(result) == 0 {
		result = append(result, map[string]interface{}{
			"deviceId": "default",
			"label":    "Default Microphone",
			"kind":     "audioinput",
			"index":    0,
		})
	}

	return result, nil
}

// StartRecording begins audio recording
func (ar *AudioRecorder) StartRecording(metadata RecordingMetadata) error {
	if ar.isRecording {
		return fmt.Errorf("already recording")
	}

	// Create recordings directory if it doesn't exist
	recordingsDir := filepath.Join("recordings")
	if err := os.MkdirAll(recordingsDir, 0755); err != nil {
		return fmt.Errorf("failed to create recordings directory: %w", err)
	}

	// Generate filename based on timestamp and lecture title
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	safeTitle := sanitizeFilename(metadata.LectureTitle)
	filename := fmt.Sprintf("%s_%s.wav", timestamp, safeTitle)
	filePath := filepath.Join(recordingsDir, filename)

	// Create output file
	outputFile, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}

	// Configure audio format
	config := malgo.DefaultDeviceConfig(malgo.Capture)
	config.Capture.Format = malgo.FormatS16
	config.Capture.Channels = 1
	config.SampleRate = 16000

	// Create device callbacks
	callbacks := malgo.DeviceCallbacks{
		Data: ar.onSamples,
	}

	// Create device
	device, err := malgo.InitDevice(ar.context.Context, config, callbacks)
	if err != nil {
		outputFile.Close()
		return fmt.Errorf("failed to create audio device: %w", err)
	}

	// Start the device
	if err := device.Start(); err != nil {
		outputFile.Close()
		return fmt.Errorf("failed to start audio device: %w", err)
	}

	ar.device = device
	ar.outputFile = outputFile
	ar.isRecording = true
	ar.stopChan = make(chan struct{})

	// Write WAV header
	if err := ar.writeWAVHeader(); err != nil {
		ar.StopRecording()
		return fmt.Errorf("failed to write WAV header: %w", err)
	}

	log.Printf("Started recording to: %s", filePath)
	return nil
}

// StopRecording stops the current recording
func (ar *AudioRecorder) StopRecording() (*RecordingMetadata, error) {
	if !ar.isRecording {
		return nil, fmt.Errorf("not currently recording")
	}

	// Stop the device
	if ar.device != nil {
		ar.device.Uninit()
	}

	// Update WAV header with final file size
	if ar.outputFile != nil {
		if err := ar.updateWAVHeader(); err != nil {
			log.Printf("Warning: failed to update WAV header: %v", err)
		}
	}

	// Close the output file
	if ar.outputFile != nil {
		ar.outputFile.Close()
	}

	ar.isRecording = false
	close(ar.stopChan)

	// Return metadata
	metadata := &RecordingMetadata{
		CourseID:     "", // Will be set by caller
		LectureTitle: "", // Will be set by caller
		Duration:     0,  // Will be calculated
		Timestamp:    time.Now(),
		FilePath:     ar.outputFile.Name(),
		SampleRate:   16000,
		Channels:     1,
		Format:       "WAV",
	}

	log.Printf("Stopped recording: %s", metadata.FilePath)

	return metadata, nil
}

// IsRecording returns whether the recorder is currently recording
func (ar *AudioRecorder) IsRecording() bool {
	return ar.isRecording
}

// onSamples is called by malgo when audio samples are available
func (ar *AudioRecorder) onSamples(out, in []byte, framecount uint32) {
	if ar.isRecording && ar.outputFile != nil {
		// Write audio data to file
		if _, err := ar.outputFile.Write(in); err != nil {
			log.Printf("Error writing audio data: %v", err)
		}
	}
}

// writeWAVHeader writes the WAV file header
func (ar *AudioRecorder) writeWAVHeader() error {
	// WAV header structure
	header := make([]byte, 44)

	// RIFF header
	copy(header[0:4], []byte("RIFF"))
	// File size (will be updated later)
	copy(header[4:8], []byte{0, 0, 0, 0})
	copy(header[8:12], []byte("WAVE"))

	// fmt chunk
	copy(header[12:16], []byte("fmt "))
	// fmt chunk size
	copy(header[16:20], []byte{16, 0, 0, 0})
	// Audio format (PCM = 1)
	copy(header[20:22], []byte{1, 0})
	// Number of channels
	copy(header[22:24], []byte{1, 0})
	// Sample rate
	copy(header[24:28], []byte{64, 61, 0, 0}) // 16000 in little endian
	// Byte rate
	copy(header[28:32], []byte{32, 122, 0, 0}) // 16000 * 2 bytes per sample
	// Block align
	copy(header[32:34], []byte{2, 0})
	// Bits per sample
	copy(header[34:36], []byte{16, 0})

	// data chunk
	copy(header[36:40], []byte("data"))
	// data chunk size (will be updated later)
	copy(header[40:44], []byte{0, 0, 0, 0})

	_, err := ar.outputFile.Write(header)
	return err
}

// updateWAVHeader updates the WAV header with the final file size
func (ar *AudioRecorder) updateWAVHeader() error {
	// Get current file size
	fileInfo, err := ar.outputFile.Stat()
	if err != nil {
		return err
	}

	fileSize := fileInfo.Size()

	// Seek to beginning of file
	if _, err := ar.outputFile.Seek(0, io.SeekStart); err != nil {
		return err
	}

	// Update file size in RIFF header
	fileSizeBytes := make([]byte, 4)
	fileSizeBytes[0] = byte(fileSize - 8)
	fileSizeBytes[1] = byte((fileSize - 8) >> 8)
	fileSizeBytes[2] = byte((fileSize - 8) >> 16)
	fileSizeBytes[3] = byte((fileSize - 8) >> 24)

	if _, err := ar.outputFile.WriteAt(fileSizeBytes, 4); err != nil {
		return err
	}

	// Update data chunk size
	dataSize := fileSize - 44
	dataSizeBytes := make([]byte, 4)
	dataSizeBytes[0] = byte(dataSize)
	dataSizeBytes[1] = byte(dataSize >> 8)
	dataSizeBytes[2] = byte(dataSize >> 16)
	dataSizeBytes[3] = byte(dataSize >> 24)

	if _, err := ar.outputFile.WriteAt(dataSizeBytes, 40); err != nil {
		return err
	}

	return nil
}

// Close cleans up the audio recorder
func (ar *AudioRecorder) Close() error {
	if ar.isRecording {
		ar.StopRecording()
	}

	if ar.context != nil {
		ar.context.Uninit()
		ar.context.Free()
	}

	return nil
}

// sanitizeFilename removes invalid characters from filename
func sanitizeFilename(filename string) string {
	// Replace invalid characters with underscores
	invalidChars := []rune{'/', '\\', ':', '*', '?', '"', '<', '>', '|'}
	result := filename

	for _, char := range invalidChars {
		result = strings.ReplaceAll(result, string(char), "_")
	}

	// Limit length
	if len(result) > 50 {
		result = result[:50]
	}

	return result
}
