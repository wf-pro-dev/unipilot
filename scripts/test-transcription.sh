#!/bin/bash

# Test script for the complete audio recording and transcription workflow
# This script will help you test the Hugging Face API integration

echo "🎤 Unipilot Audio Transcription Test"
echo "====================================="
echo ""

# Check if HUGGINGFACE_API_KEY is set
if [ -z "$HUGGINGFACE_API_KEY" ]; then
    echo "❌ Error: HUGGINGFACE_API_KEY environment variable is not set"
    echo "Please set it with: export HUGGINGFACE_API_KEY='your_api_key_here'"
    exit 1
fi

echo "✅ Hugging Face API key is set"
echo ""

# Check if the app is built
if [ ! -f "./unipilot" ]; then
    echo "🔨 Building the application..."
    go build -o unipilot .
    if [ $? -ne 0 ]; then
        echo "❌ Build failed"
        exit 1
    fi
fi

echo "✅ Application is built"
echo ""

echo "📋 Test Instructions:"
echo "1. The app will start in development mode"
echo "2. Go to the Notes section in the frontend"
echo "3. Select a course and start recording"
echo "4. Speak for 10-30 seconds (test content)"
echo "5. Stop the recording"
echo "6. The transcription will be processed via Hugging Face API"
echo "7. Check the terminal for logs and results"
echo ""

echo "🚀 Starting the application..."
echo "Press Ctrl+C to stop the test"
echo ""

# Start the application
./unipilot 