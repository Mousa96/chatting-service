package integration

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileUpload(t *testing.T) {
	// Setup test server and user
	userToken := setupTestUser("uploaduser", "pass123")

	tests := []struct {
		name         string
		fileContent  []byte
		filename     string
		contentType  string
		expectedCode int
	}{
		{
			name:         "Valid image upload",
			fileContent:  []byte{0xFF, 0xD8, 0xFF, 0xE0}, // JPEG header
			filename:     "test.jpg",
			contentType:  "image/jpeg",
			expectedCode: http.StatusOK,
		},
		{
			name:         "Invalid file type",
			fileContent:  []byte("text content"),
			filename:     "test.txt",
			contentType:  "text/plain",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "Empty file",
			fileContent:  []byte{},
			filename:     "empty.jpg",
			contentType:  "image/jpeg",
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)

			part, err := writer.CreateFormFile("file", tt.filename)
			require.NoError(t, err)

			_, err = part.Write(tt.fileContent)
			require.NoError(t, err)

			err = writer.Close()
			require.NoError(t, err)

			// Use httptest.NewRequest and testServer.ServeHTTP
			req := httptest.NewRequest(http.MethodPost, "/api/messages/upload", body)
			req.Header.Set("Content-Type", writer.FormDataContentType())
			req.Header.Set("Authorization", "Bearer "+userToken)

			rr := httptest.NewRecorder()
			testServer.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedCode, rr.Code)

			if tt.expectedCode == http.StatusOK {
				var response struct {
					URL string `json:"url"`
				}
				err = json.NewDecoder(rr.Body).Decode(&response)
				require.NoError(t, err)
				assert.Contains(t, response.URL, "uploads/")
			}
		})
	}
}

func TestFileServing(t *testing.T) {
	// Create test uploads directory if it doesn't exist
	uploadDir := "/app/uploads" // Match the directory used in Docker
	err := os.MkdirAll(uploadDir, 0755)
	require.NoError(t, err)
	defer os.RemoveAll(uploadDir) // Clean up after test
	
	// Create a test file
	testContent := []byte("test content")
	filename := "test_serve.txt"
	uploadPath := filepath.Join(uploadDir, filename)
	
	err = os.WriteFile(uploadPath, testContent, 0644)
	require.NoError(t, err)

	// Test file serving
	req := httptest.NewRequest(http.MethodGet, "/uploads/"+filename, nil)
	rr := httptest.NewRecorder()
	
	// Create new file server handler
	fs := http.StripPrefix("/uploads/", http.FileServer(http.Dir(uploadDir)))
	fs.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, testContent, rr.Body.Bytes())
} 