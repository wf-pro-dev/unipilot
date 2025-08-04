package audio

import (
	"os"
	"testing"
	"time"
)

func TestAudioRecorder(t *testing.T) {
	// Test creating audio recorder
	recorder, err := NewAudioRecorder()
	if err != nil {
		t.Fatalf("Failed to create audio recorder: %v", err)
	}
	defer recorder.Close()

	// Test getting audio devices
	devices, err := recorder.GetAudioDevices()
	if err != nil {
		t.Fatalf("Failed to get audio devices: %v", err)
	}

	t.Logf("Found %d audio devices", len(devices))
	for i, device := range devices {
		t.Logf("Device %d: %s (%s)", i+1, device["label"], device["deviceId"])
	}

	// Test recording functionality (this may fail if no microphone is available)
	metadata := RecordingMetadata{
		CourseID:     "test-course",
		LectureTitle: "Test Recording",
		Timestamp:    time.Now(),
	}

	err = recorder.StartRecording(metadata)
	if err != nil {
		t.Logf("Recording test failed (expected if no microphone): %v", err)
		return
	}

	// Record for 1 second
	time.Sleep(1 * time.Second)

	// Stop recording
	result, err := recorder.StopRecording()
	if err != nil {
		t.Fatalf("Failed to stop recording: %v", err)
	}

	t.Logf("Recording completed: %s", result.FilePath)

	// Check if file was created
	if _, err := os.Stat(result.FilePath); os.IsNotExist(err) {
		t.Errorf("Recording file was not created: %s", result.FilePath)
	} else {
		t.Logf("Recording file created successfully: %s", result.FilePath)
	}
}

func TestAudioRecorderPermissions(t *testing.T) {
	recorder, err := NewAudioRecorder()
	if err != nil {
		t.Fatalf("Failed to create audio recorder: %v", err)
	}
	defer recorder.Close()

	// Test permissions check
	devices, err := recorder.GetAudioDevices()
	if err != nil {
		t.Logf("Permission test: %v", err)
	} else {
		t.Logf("Permission test passed, found %d devices", len(devices))
	}
}
