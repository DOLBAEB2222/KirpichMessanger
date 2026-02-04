package tests

import (
    "bytes"
    "image"
    "image/color"
    "image/jpeg"
    "image/png"
    "mime/multipart"
    "os"
    "path/filepath"
    "strings"
    "testing"

    "github.com/messenger/backend/pkg/media"
)

func createTestImage(t *testing.T, format string, width, height int) (string, int64) {
    tmpDir := t.TempDir()
    
    img := image.NewRGBA(image.Rect(0, 0, width, height))
    
    for y := 0; y < height; y++ {
        for x := 0; x < width; x++ {
            img.Set(x, y, color.RGBA{uint8(x % 256), uint8(y % 256), 128, 255})
        }
    }
    
    var buf bytes.Buffer
    var err error
    var filename string
    
    switch format {
    case "jpeg", "jpg":
        err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 80})
        filename = "test.jpg"
    case "png":
        err = png.Encode(&buf, img)
        filename = "test.png"
    default:
        err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 80})
        filename = "test.jpg"
    }
    
    if err != nil {
        t.Fatalf("failed to encode test image: %v", err)
    }
    
    filePath := filepath.Join(tmpDir, filename)
    err = os.WriteFile(filePath, buf.Bytes(), 0644)
    if err != nil {
        t.Fatalf("failed to write test image: %v", err)
    }
    
    info, _ := os.Stat(filePath)
    return filePath, info.Size()
}

func TestMediaUploader_ValidateFile(t *testing.T) {
    uploader := media.NewMediaUploader()
    
    tests := []struct {
        name         string
        filename     string
        contentType  string
        size         int64
        allowedTypes map[string]bool
        wantErr      bool
        errContains  string
    }{
        {
            name:         "Valid JPEG image",
            filename:     "test.jpg",
            contentType:  "image/jpeg",
            size:         1024 * 1024,
            allowedTypes: media.AllowedImageTypes,
            wantErr:      false,
        },
        {
            name:         "Valid PNG image",
            filename:     "test.png",
            contentType:  "image/png",
            size:         1024 * 1024,
            allowedTypes: media.AllowedImageTypes,
            wantErr:      false,
        },
        {
            name:         "Invalid file type",
            filename:     "test.exe",
            contentType:  "application/octet-stream",
            size:         1024,
            allowedTypes: media.AllowedImageTypes,
            wantErr:      true,
            errContains:  "not allowed",
        },
        {
            name:         "File too large",
            filename:     "large.jpg",
            contentType:  "image/jpeg",
            size:         media.MaxFileSize + 1,
            allowedTypes: media.AllowedImageTypes,
            wantErr:      true,
            errContains:  "exceeds",
        },
        {
            name:         "Invalid filename with path traversal",
            filename:     "../etc/passwd",
            contentType:  "image/jpeg",
            size:         1024,
            allowedTypes: media.AllowedImageTypes,
            wantErr:      true,
            errContains:  "invalid filename",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            header := &multipart.FileHeader{
                Filename: tt.filename,
                Size:     tt.size,
                Header:   map[string][]string{"Content-Type": {tt.contentType}},
            }
            
            err := uploader.ValidateFile(header, tt.allowedTypes)
            if tt.wantErr {
                if err == nil {
                    t.Errorf("expected error but got none")
                } else if !strings.Contains(err.Error(), tt.errContains) {
                    t.Errorf("expected error containing %q, got %q", tt.errContains, err.Error())
                }
            } else {
                if err != nil {
                    t.Errorf("unexpected error: %v", err)
                }
            }
        })
    }
}

func TestGetImageDimensions(t *testing.T) {
    tests := []struct {
        name          string
        format        string
        width         int
        height        int
        wantWidth     int
        wantHeight    int
        wantErr       bool
    }{
        {
            name:       "JPEG image dimensions",
            format:     "jpeg",
            width:      800,
            height:     600,
            wantWidth:  800,
            wantHeight: 600,
            wantErr:    false,
        },
        {
            name:       "PNG image dimensions",
            format:     "png",
            width:      1920,
            height:     1080,
            wantWidth:  1920,
            wantHeight: 1080,
            wantErr:    false,
        },
        {
            name:       "Square image",
            format:     "jpeg",
            width:      500,
            height:     500,
            wantWidth:  500,
            wantHeight: 500,
            wantErr:    false,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            filePath, _ := createTestImage(t, tt.format, tt.width, tt.height)
            
            gotWidth, gotHeight, err := media.GetImageDimensions(filePath)
            if tt.wantErr {
                if err == nil {
                    t.Error("expected error but got none")
                }
                return
            }
            
            if err != nil {
                t.Errorf("unexpected error: %v", err)
                return
            }
            
            if gotWidth != tt.wantWidth {
                t.Errorf("width = %d, want %d", gotWidth, tt.wantWidth)
            }
            if gotHeight != tt.wantHeight {
                t.Errorf("height = %d, want %d", gotHeight, tt.wantHeight)
            }
        })
    }
}

func TestCompressImage(t *testing.T) {
    tmpDir := t.TempDir()
    
    tests := []struct {
        name           string
        format         string
        inputWidth     int
        inputHeight    int
        maxWidth       int
        expectResize   bool
    }{
        {
            name:         "Resize large image",
            format:       "jpeg",
            inputWidth:   2000,
            inputHeight:  1500,
            maxWidth:     500,
            expectResize: true,
        },
        {
            name:         "No resize for small image",
            format:       "jpeg",
            inputWidth:   400,
            inputHeight:  300,
            maxWidth:     500,
            expectResize: false,
        },
        {
            name:         "Resize exact width",
            format:       "png",
            inputWidth:   1000,
            inputHeight:  750,
            maxWidth:     500,
            expectResize: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            inputPath, _ := createTestImage(t, tt.format, tt.inputWidth, tt.inputHeight)
            outputPath := filepath.Join(tmpDir, tt.name+"_compressed.jpg")
            
            err := media.CompressImage(inputPath, outputPath, tt.maxWidth)
            if err != nil {
                t.Fatalf("CompressImage failed: %v", err)
            }
            
            if _, err := os.Stat(outputPath); os.IsNotExist(err) {
                t.Fatal("compressed file was not created")
            }
            
            width, height, err := media.GetImageDimensions(outputPath)
            if err != nil {
                t.Fatalf("failed to get compressed image dimensions: %v", err)
            }
            
            if tt.expectResize {
                if width > tt.maxWidth {
                    t.Errorf("width %d exceeds maxWidth %d", width, tt.maxWidth)
                }
                
                expectedHeight := (tt.inputHeight * tt.maxWidth) / tt.inputWidth
                if height != expectedHeight {
                    t.Errorf("height = %d, want %d", height, expectedHeight)
                }
            }
        })
    }
}

func TestGenerateThumbnail(t *testing.T) {
    tmpDir := t.TempDir()
    
    inputPath, _ := createTestImage(t, "jpeg", 800, 600)
    outputPath := filepath.Join(tmpDir, "thumb.jpg")
    maxSize := 200
    
    err := media.GenerateThumbnail(inputPath, outputPath, maxSize)
    if err != nil {
        t.Fatalf("GenerateThumbnail failed: %v", err)
    }
    
    if _, err := os.Stat(outputPath); os.IsNotExist(err) {
        t.Fatal("thumbnail was not created")
    }
    
    width, height, err := media.GetImageDimensions(outputPath)
    if err != nil {
        t.Fatalf("failed to get thumbnail dimensions: %v", err)
    }
    
    if width > maxSize && height > maxSize {
        t.Errorf("thumbnail %dx%d exceeds max size %d", width, height, maxSize)
    }
}

func TestCalculateOptimalQuality(t *testing.T) {
    tests := []struct {
        fileSize int64
        want     float32
    }{
        {fileSize: 15 * 1024 * 1024, want: 70},
        {fileSize: 10 * 1024 * 1024, want: 70},
        {fileSize: 7 * 1024 * 1024, want: 75},
        {fileSize: 5 * 1024 * 1024, want: 75},
        {fileSize: 3 * 1024 * 1024, want: 80},
        {fileSize: 1 * 1024 * 1024, want: 80},
        {fileSize: 500 * 1024, want: 85},
    }
    
    for _, tt := range tests {
        t.Run(string(rune(tt.fileSize)), func(t *testing.T) {
            got := media.CalculateOptimalQuality(tt.fileSize)
            if got != tt.want {
                t.Errorf("CalculateOptimalQuality(%d) = %f, want %f", tt.fileSize, got, tt.want)
            }
        })
    }
}

func TestAllowedTypes(t *testing.T) {
    tests := []struct {
        mimeType   string
        isImage    bool
        isVideo    bool
        isAudio    bool
    }{
        {"image/jpeg", true, false, false},
        {"image/png", true, false, false},
        {"image/gif", true, false, false},
        {"image/webp", true, false, false},
        {"video/mp4", false, true, false},
        {"video/webm", false, true, false},
        {"audio/mpeg", false, false, true},
        {"audio/mp3", false, false, true},
        {"application/pdf", false, false, false},
    }
    
    for _, tt := range tests {
        t.Run(tt.mimeType, func(t *testing.T) {
            isImage := media.AllowedImageTypes[tt.mimeType]
            isVideo := media.AllowedVideoTypes[tt.mimeType]
            isAudio := media.AllowedAudioTypes[tt.mimeType]
            
            if isImage != tt.isImage {
                t.Errorf("image: got %v, want %v", isImage, tt.isImage)
            }
            if isVideo != tt.isVideo {
                t.Errorf("video: got %v, want %v", isVideo, tt.isVideo)
            }
            if isAudio != tt.isAudio {
                t.Errorf("audio: got %v, want %v", isAudio, tt.isAudio)
            }
        })
    }
}
