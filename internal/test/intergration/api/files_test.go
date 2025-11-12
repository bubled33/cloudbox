package api_test

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourusername/cloud-file-storage/internal/domain/file_version"
)

func TestFiles_GetAll_Success(t *testing.T) {
	env, cleanup := SetupTestEnvironment(t)
	defer cleanup()

	accessToken, _ := createUserAndGetTokens(t, env, "test@mail.ru", "test")

	// Создаем несколько файлов (без реальной загрузки)
	fileID1 := createFileWithStatus(t, env, "document.pdf", 1024, "application/pdf", accessToken, file_version.FileStatusReady)
	fileID2 := createFileWithStatus(t, env, "image.jpg", 2048, "image/jpeg", accessToken, file_version.FileStatusReady)

	w := env.NewRequestWithAuth(t, "GET", "/api/v1/files", nil, accessToken)

	require.Equal(t, 200, w.Code)

	response := ParseJSONResponse(t, w)

	require.Contains(t, response, "files")
	require.Contains(t, response, "total")

	files, ok := response["files"].([]interface{})
	require.True(t, ok, "files should be an array")
	require.GreaterOrEqual(t, len(files), 2, "should have at least 2 files")

	firstFile, ok := files[0].(map[string]interface{})
	require.True(t, ok)
	assert.Contains(t, firstFile, "id")
	assert.Contains(t, firstFile, "name")
	assert.Contains(t, firstFile, "size")
	assert.Contains(t, firstFile, "mime")
	assert.Equal(t, "ready", firstFile["status"])

	_ = fileID1
	_ = fileID2
}

func TestFiles_GetAll_WithPagination(t *testing.T) {
	env, cleanup := SetupTestEnvironment(t)
	defer cleanup()

	accessToken, _ := createUserAndGetTokens(t, env, "test@mail.ru", "test")

	// Создаем 5 файлов
	for i := 0; i < 5; i++ {
		createFileWithStatus(t, env, fmt.Sprintf("file%d.txt", i), uint64(100*(i+1)), "text/plain", accessToken, file_version.FileStatusReady)
	}

	w := env.NewRequestWithAuth(t, "GET", "/api/v1/files?limit=2", nil, accessToken)
	require.Equal(t, 200, w.Code)

	response := ParseJSONResponse(t, w)
	files := response["files"].([]interface{})
	assert.LessOrEqual(t, len(files), 2)
	assert.Equal(t, float64(2), response["limit"])
}

func TestFiles_GetAll_WithSearch(t *testing.T) {
	env, cleanup := SetupTestEnvironment(t)
	defer cleanup()

	accessToken, _ := createUserAndGetTokens(t, env, "test@mail.ru", "test")

	createFileWithStatus(t, env, "report.pdf", 1024, "application/pdf", accessToken, file_version.FileStatusReady)
	createFileWithStatus(t, env, "image.jpg", 2048, "image/jpeg", accessToken, file_version.FileStatusReady)

	w := env.NewRequestWithAuth(t, "GET", "/api/v1/files?q=report", nil, accessToken)
	require.Equal(t, 200, w.Code)

	response := ParseJSONResponse(t, w)
	files := response["files"].([]interface{})
	require.Greater(t, len(files), 0)

	firstFile := files[0].(map[string]interface{})
	assert.Contains(t, firstFile["name"], "report")
}

func TestFiles_GetAll_Unauthorized(t *testing.T) {
	env, cleanup := SetupTestEnvironment(t)
	defer cleanup()

	w := env.NewRequest(t, "GET", "/api/v1/files", nil)
	require.Equal(t, 401, w.Code)
}

func TestFiles_GetById_Success(t *testing.T) {
	env, cleanup := SetupTestEnvironment(t)
	defer cleanup()

	accessToken, _ := createUserAndGetTokens(t, env, "test@mail.ru", "test")
	fileID := createFileWithStatus(t, env, "test.pdf", 128, "application/pdf", accessToken, file_version.FileStatusReady)

	w := env.NewRequestWithAuth(t, "GET", "/api/v1/files/"+fileID.String(), nil, accessToken)

	require.Equal(t, 200, w.Code)

	response := ParseJSONResponse(t, w)

	assert.Equal(t, fileID.String(), response["id"])
	assert.Equal(t, "test.pdf", response["name"])
	assert.Equal(t, "application/pdf", response["mime"])
	assert.Equal(t, "ready", response["status"])
	assert.NotEmpty(t, response["owner_id"])
	assert.NotEmpty(t, response["created_at"])
	assert.Contains(t, response, "total_versions")
}

func TestFiles_GetById_NotFound(t *testing.T) {
	env, cleanup := SetupTestEnvironment(t)
	defer cleanup()

	accessToken, _ := createUserAndGetTokens(t, env, "test@mail.ru", "test")

	fakeID := "00000000-0000-0000-0000-000000000000"
	w := env.NewRequestWithAuth(t, "GET", "/api/v1/files/"+fakeID, nil, accessToken)

	require.Equal(t, 404, w.Code)

	response := ParseJSONResponse(t, w)
	// Изменено: проверяем "code" вместо "error"
	assert.Contains(t, response, "code")
	assert.Equal(t, "FILE_NOT_FOUND", response["code"])
}

func TestFiles_GetById_InvalidUUID(t *testing.T) {
	env, cleanup := SetupTestEnvironment(t)
	defer cleanup()

	accessToken, _ := createUserAndGetTokens(t, env, "test@mail.ru", "test")

	w := env.NewRequestWithAuth(t, "GET", "/api/v1/files/invalid-uuid", nil, accessToken)

	require.Equal(t, 400, w.Code)

	response := ParseJSONResponse(t, w)
	assert.Equal(t, "invalid file_id format", response["error"])
}

func TestFiles_GetById_AccessDenied(t *testing.T) {
	env, cleanup := SetupTestEnvironment(t)
	defer cleanup()

	token1, _ := createUserAndGetTokens(t, env, "user1@mail.ru", "User 1")
	fileID := createFileWithStatus(t, env, "private.pdf", 128, "application/pdf", token1, file_version.FileStatusReady)

	token2, _ := createUserAndGetTokens(t, env, "user2@mail.ru", "User 2")
	w := env.NewRequestWithAuth(t, "GET", "/api/v1/files/"+fileID.String(), nil, token2)

	require.Equal(t, 403, w.Code)

	response := ParseJSONResponse(t, w)
	assert.Equal(t, "access denied", response["error"])
}

func TestFiles_Update_Success(t *testing.T) {
	env, cleanup := SetupTestEnvironment(t)
	defer cleanup()

	accessToken, _ := createUserAndGetTokens(t, env, "test@mail.ru", "test")
	fileID := createFileWithStatus(t, env, "old_name.pdf", 128, "application/pdf", accessToken, file_version.FileStatusReady)

	body := map[string]interface{}{
		"name": "new_name.pdf",
	}
	w := env.NewJSONRequestWithAuth(t, "PATCH", "/api/v1/files/"+fileID.String(), body, accessToken)

	require.Equal(t, 200, w.Code)

	response := ParseJSONResponse(t, w)
	assert.Equal(t, "new_name.pdf", response["name"])
	assert.Equal(t, fileID.String(), response["id"])
}

func TestFiles_Update_InvalidInput(t *testing.T) {
	env, cleanup := SetupTestEnvironment(t)
	defer cleanup()

	accessToken, _ := createUserAndGetTokens(t, env, "test@mail.ru", "test")
	fileID := createFileWithStatus(t, env, "test.pdf", 128, "application/pdf", accessToken, file_version.FileStatusReady)

	tests := []struct {
		name string
		body map[string]interface{}
	}{
		{
			name: "empty name",
			body: map[string]interface{}{
				"name": "",
			},
		},
		{
			name: "missing name",
			body: map[string]interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := env.NewJSONRequestWithAuth(t, "PATCH", "/api/v1/files/"+fileID.String(), tt.body, accessToken)
			require.Equal(t, 400, w.Code)

			response := ParseJSONResponse(t, w)
			assert.Contains(t, response, "error")
		})
	}
}

func TestFiles_Update_AccessDenied(t *testing.T) {
	env, cleanup := SetupTestEnvironment(t)
	defer cleanup()

	token1, _ := createUserAndGetTokens(t, env, "user1@mail.ru", "User 1")
	fileID := createFileWithStatus(t, env, "protected.pdf", 128, "application/pdf", token1, file_version.FileStatusReady)

	token2, _ := createUserAndGetTokens(t, env, "user2@mail.ru", "User 2")

	body := map[string]interface{}{
		"name": "hacked.pdf",
	}
	w := env.NewJSONRequestWithAuth(t, "PATCH", "/api/v1/files/"+fileID.String(), body, token2)

	require.Equal(t, 403, w.Code)
}

func TestFiles_Delete_Success(t *testing.T) {
	env, cleanup := SetupTestEnvironment(t)
	defer cleanup()

	accessToken, _ := createUserAndGetTokens(t, env, "test@mail.ru", "test")
	fileID := createFileWithStatus(t, env, "to_delete.pdf", 128, "application/pdf", accessToken, file_version.FileStatusReady)

	w := env.NewRequestWithAuth(t, "GET", "/api/v1/files/"+fileID.String(), nil, accessToken)
	require.Equal(t, 200, w.Code)

	w = env.NewRequestWithAuth(t, "DELETE", "/api/v1/files/"+fileID.String(), nil, accessToken)
	require.Equal(t, 200, w.Code)

	response := ParseJSONResponse(t, w)
	assert.Contains(t, response, "message")
	assert.Equal(t, "File deleted successfully", response["message"])

	w = env.NewRequestWithAuth(t, "GET", "/api/v1/files/"+fileID.String(), nil, accessToken)
	require.Equal(t, 404, w.Code)
}

func TestFiles_Delete_AccessDenied(t *testing.T) {
	env, cleanup := SetupTestEnvironment(t)
	defer cleanup()

	token1, _ := createUserAndGetTokens(t, env, "user1@mail.ru", "User 1")
	fileID := createFileWithStatus(t, env, "protected.pdf", 128, "application/pdf", token1, file_version.FileStatusReady)

	token2, _ := createUserAndGetTokens(t, env, "user2@mail.ru", "User 2")
	w := env.NewRequestWithAuth(t, "DELETE", "/api/v1/files/"+fileID.String(), nil, token2)

	require.Equal(t, 403, w.Code)
}

func TestFiles_Delete_NotFound(t *testing.T) {
	env, cleanup := SetupTestEnvironment(t)
	defer cleanup()

	accessToken, _ := createUserAndGetTokens(t, env, "test@mail.ru", "test")

	fakeID := uuid.New()
	w := env.NewRequestWithAuth(t, "DELETE", "/api/v1/files/"+fakeID.String(), nil, accessToken)

	require.Equal(t, 404, w.Code)
}

func TestFiles_UploadNewVersion_Success(t *testing.T) {
	env, cleanup := SetupTestEnvironment(t)
	defer cleanup()

	accessToken, _ := createUserAndGetTokens(t, env, "test@mail.ru", "test")
	// Используем createFileAndUpload для реального теста загрузки
	fileID := createFileAndUpload(t, env, "document.pdf", 1024, "application/pdf", accessToken, file_version.FileStatusReady)

	body := map[string]interface{}{
		"name": "document.pdf",
		"size": 2048,
		"mime": "application/pdf",
	}
	w := env.NewJSONRequestWithAuth(t, "POST", "/api/v1/files/"+fileID.String()+"/versions", body, accessToken)

	require.Equal(t, 201, w.Code)

	response := ParseJSONResponse(t, w)
	assert.Equal(t, fileID.String(), response["file_id"])
	assert.Contains(t, response, "version_id")
	assert.Contains(t, response, "upload_url")
	assert.Equal(t, float64(2), response["version_num"])
	assert.Equal(t, "processing", response["status"])

	uploadURL := response["upload_url"].(string)
	err := uploadFileToS3(t, uploadURL, make([]byte, 2048))
	require.NoError(t, err)

	waitForFileStatus(t, env, fileID, accessToken, file_version.FileStatusUploaded)
	waitForFileStatus(t, env, fileID, accessToken, file_version.FileStatusReady)

	w = env.NewRequestWithAuth(t, "GET", "/api/v1/files/"+fileID.String(), nil, accessToken)
	require.Equal(t, 200, w.Code)

	response = ParseJSONResponse(t, w)
	assert.Equal(t, float64(2), response["current_version"])
}

func TestFiles_UploadNewVersion_FileSizeExceeded(t *testing.T) {
	env, cleanup := SetupTestEnvironment(t)
	defer cleanup()

	accessToken, _ := createUserAndGetTokens(t, env, "test@mail.ru", "test")
	fileID := createFileWithStatus(t, env, "document.pdf", 1024, "application/pdf", accessToken, file_version.FileStatusReady)

	body := map[string]interface{}{
		"name": "huge.pdf",
		"size": uint64(6 * 1024 * 1024 * 1024),
		"mime": "application/pdf",
	}
	w := env.NewJSONRequestWithAuth(t, "POST", "/api/v1/files/"+fileID.String()+"/versions", body, accessToken)

	require.Equal(t, 400, w.Code)

	response := ParseJSONResponse(t, w)
	assert.Contains(t, response["error"], "exceeds maximum")
}

func TestFiles_UploadNewVersion_AccessDenied(t *testing.T) {
	env, cleanup := SetupTestEnvironment(t)
	defer cleanup()

	token1, _ := createUserAndGetTokens(t, env, "user1@mail.ru", "User 1")
	fileID := createFileWithStatus(t, env, "document.pdf", 1024, "application/pdf", token1, file_version.FileStatusReady)

	token2, _ := createUserAndGetTokens(t, env, "user2@mail.ru", "User 2")

	body := map[string]interface{}{
		"name": "document.pdf",
		"size": 2048,
		"mime": "application/pdf",
	}
	w := env.NewJSONRequestWithAuth(t, "POST", "/api/v1/files/"+fileID.String()+"/versions", body, token2)

	require.Equal(t, 403, w.Code)
}

func TestFiles_GetVersions_Success(t *testing.T) {
	env, cleanup := SetupTestEnvironment(t)
	defer cleanup()

	accessToken, _ := createUserAndGetTokens(t, env, "test@mail.ru", "test")
	fileID := createFileWithStatus(t, env, "versioned.pdf", 1024, "application/pdf", accessToken, file_version.FileStatusReady)

	w := env.NewRequestWithAuth(t, "GET", "/api/v1/files/"+fileID.String()+"/versions", nil, accessToken)

	require.Equal(t, 200, w.Code)

	response := ParseJSONResponse(t, w)
	assert.Contains(t, response, "versions")
	assert.Contains(t, response, "total")

	versions := response["versions"].([]interface{})
	assert.GreaterOrEqual(t, len(versions), 1)

	firstVersion := versions[0].(map[string]interface{})
	assert.Equal(t, fileID.String(), firstVersion["file_id"])
	assert.Equal(t, float64(1), firstVersion["version_num"])
	assert.Equal(t, "ready", firstVersion["status"])
}

func TestFiles_GetVersions_MultipleVersions(t *testing.T) {
	env, cleanup := SetupTestEnvironment(t)
	defer cleanup()

	accessToken, _ := createUserAndGetTokens(t, env, "test@mail.ru", "test")
	// Используем createFileAndUpload для реального теста с версиями
	fileID := createFileAndUpload(t, env, "versioned.pdf", 1024, "application/pdf", accessToken, file_version.FileStatusReady)

	body := map[string]interface{}{
		"name": "versioned.pdf",
		"size": 2048,
		"mime": "application/pdf",
	}
	w := env.NewJSONRequestWithAuth(t, "POST", "/api/v1/files/"+fileID.String()+"/versions", body, accessToken)
	require.Equal(t, 201, w.Code)

	response := ParseJSONResponse(t, w)
	uploadURL := response["upload_url"].(string)
	err := uploadFileToS3(t, uploadURL, make([]byte, 2048))
	require.NoError(t, err)

	waitForFileStatus(t, env, fileID, accessToken, file_version.FileStatusUploaded)
	waitForFileStatus(t, env, fileID, accessToken, file_version.FileStatusReady)

	w = env.NewRequestWithAuth(t, "GET", "/api/v1/files/"+fileID.String()+"/versions", nil, accessToken)
	require.Equal(t, 200, w.Code)

	response = ParseJSONResponse(t, w)
	versions := response["versions"].([]interface{})
	assert.Equal(t, 2, len(versions))
	assert.Equal(t, float64(2), response["total"])
}

func TestFiles_CreatePublicLink_Success(t *testing.T) {
	env, cleanup := SetupTestEnvironment(t)
	defer cleanup()

	accessToken, _ := createUserAndGetTokens(t, env, "test@mail.ru", "test")
	fileID := createFileWithStatus(t, env, "shared.pdf", 1024, "application/pdf", accessToken, file_version.FileStatusReady)

	body := map[string]interface{}{
		"expires_in": "24h",
	}
	w := env.NewJSONRequestWithAuth(t, "POST", "/api/v1/files/"+fileID.String()+"/public-links", body, accessToken)

	require.Equal(t, 201, w.Code)

	response := ParseJSONResponse(t, w)
	assert.Equal(t, fileID.String(), response["file_id"])
	assert.Contains(t, response, "token")
	assert.Contains(t, response, "expires_at")
	assert.Contains(t, response, "id")
	assert.NotEmpty(t, response["token"])
	assert.Equal(t, false, response["is_expired"])
}

func TestFiles_CreatePublicLink_InvalidExpiration(t *testing.T) {
	env, cleanup := SetupTestEnvironment(t)
	defer cleanup()

	accessToken, _ := createUserAndGetTokens(t, env, "test@mail.ru", "test")
	fileID := createFileWithStatus(t, env, "shared.pdf", 1024, "application/pdf", accessToken, file_version.FileStatusReady)

	body := map[string]interface{}{
		"expires_in": "invalid-duration",
	}
	w := env.NewJSONRequestWithAuth(t, "POST", "/api/v1/files/"+fileID.String()+"/public-links", body, accessToken)

	require.Equal(t, 400, w.Code)

	response := ParseJSONResponse(t, w)
	assert.Contains(t, response["error"], "invalid expires_in format")
}

func TestFiles_GetPublicLinks_Success(t *testing.T) {
	env, cleanup := SetupTestEnvironment(t)
	defer cleanup()

	accessToken, _ := createUserAndGetTokens(t, env, "test@mail.ru", "test")
	fileID := createFileWithStatus(t, env, "shared.pdf", 1024, "application/pdf", accessToken, file_version.FileStatusReady)

	body := map[string]interface{}{
		"expires_in": "1h",
	}
	w := env.NewJSONRequestWithAuth(t, "POST", "/api/v1/files/"+fileID.String()+"/public-links", body, accessToken)
	require.Equal(t, 201, w.Code)

	w = env.NewRequestWithAuth(t, "GET", "/api/v1/files/"+fileID.String()+"/public-links", nil, accessToken)

	require.Equal(t, 200, w.Code)

	response := ParseJSONResponse(t, w)
	assert.Contains(t, response, "links")
	assert.Contains(t, response, "total")

	links := response["links"].([]interface{})
	assert.GreaterOrEqual(t, len(links), 1)

	firstLink := links[0].(map[string]interface{})
	assert.Equal(t, fileID.String(), firstLink["file_id"])
	assert.Contains(t, firstLink, "token")
}

func TestFiles_DeletePublicLink_Success(t *testing.T) {
	env, cleanup := SetupTestEnvironment(t)
	defer cleanup()

	accessToken, _ := createUserAndGetTokens(t, env, "test@mail.ru", "test")
	fileID := createFileWithStatus(t, env, "shared.pdf", 1024, "application/pdf", accessToken, file_version.FileStatusReady)

	body := map[string]interface{}{
		"expires_in": "1h",
	}
	w := env.NewJSONRequestWithAuth(t, "POST", "/api/v1/files/"+fileID.String()+"/public-links", body, accessToken)
	require.Equal(t, 201, w.Code)

	linkResponse := ParseJSONResponse(t, w)
	linkIDStr, ok := linkResponse["id"].(string)
	require.True(t, ok, "link id should be a string")

	w = env.NewRequestWithAuth(t, "DELETE", "/api/v1/files/"+fileID.String()+"/public-links/"+linkIDStr, nil, accessToken)
	require.Equal(t, 200, w.Code)

	response := ParseJSONResponse(t, w)
	assert.Contains(t, response, "message")

	w = env.NewRequestWithAuth(t, "GET", "/api/v1/files/"+fileID.String()+"/public-links", nil, accessToken)
	require.Equal(t, 200, w.Code)

	response = ParseJSONResponse(t, w)
	links := response["links"].([]interface{})
	assert.Equal(t, 0, len(links))
}

func TestFiles_UploadNewFile_Success(t *testing.T) {
	env, cleanup := SetupTestEnvironment(t)
	defer cleanup()

	accessToken, _ := createUserAndGetTokens(t, env, "test@mail.ru", "test")

	body := map[string]interface{}{
		"name": "new_upload.pdf",
		"size": 2048,
		"mime": "application/pdf",
	}
	w := env.NewJSONRequestWithAuth(t, "POST", "/api/v1/files", body, accessToken)

	require.Equal(t, 201, w.Code)

	response := ParseJSONResponse(t, w)
	assert.Contains(t, response, "file_id")
	assert.Contains(t, response, "version_id")
	assert.Contains(t, response, "upload_url")
	assert.Equal(t, float64(1), response["version_num"])
	assert.Equal(t, "processing", response["status"])
	assert.Contains(t, response, "expires_in")
}

func TestFiles_UploadNewFile_InvalidInput(t *testing.T) {
	env, cleanup := SetupTestEnvironment(t)
	defer cleanup()

	accessToken, _ := createUserAndGetTokens(t, env, "test@mail.ru", "test")

	tests := []struct {
		name string
		body map[string]interface{}
	}{
		{
			name: "missing name",
			body: map[string]interface{}{
				"size": 1024,
				"mime": "application/pdf",
			},
		},
		{
			name: "zero size",
			body: map[string]interface{}{
				"name": "test.pdf",
				"size": 0,
				"mime": "application/pdf",
			},
		},
		{
			name: "missing mime",
			body: map[string]interface{}{
				"name": "test.pdf",
				"size": 1024,
			},
		},
		{
			name: "empty name",
			body: map[string]interface{}{
				"name": "",
				"size": 1024,
				"mime": "application/pdf",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := env.NewJSONRequestWithAuth(t, "POST", "/api/v1/files", tt.body, accessToken)
			require.Equal(t, 400, w.Code)

			response := ParseJSONResponse(t, w)
			assert.Contains(t, response, "error")
		})
	}
}

func TestFiles_UploadNewFile_SizeExceeded(t *testing.T) {
	env, cleanup := SetupTestEnvironment(t)
	defer cleanup()

	accessToken, _ := createUserAndGetTokens(t, env, "test@mail.ru", "test")

	body := map[string]interface{}{
		"name": "huge.pdf",
		"size": uint64(6 * 1024 * 1024 * 1024),
		"mime": "application/pdf",
	}
	w := env.NewJSONRequestWithAuth(t, "POST", "/api/v1/files", body, accessToken)

	require.Equal(t, 400, w.Code)

	response := ParseJSONResponse(t, w)
	assert.Contains(t, response["error"], "exceeds maximum")
}

func TestFiles_UploadNewFile_Unauthorized(t *testing.T) {
	env, cleanup := SetupTestEnvironment(t)
	defer cleanup()

	body := map[string]interface{}{
		"name": "test.pdf",
		"size": 1024,
		"mime": "application/pdf",
	}
	w := env.NewJSONRequest(t, "POST", "/api/v1/files", body)

	require.Equal(t, 401, w.Code)
}

func TestFiles_GetDownloadURL_Success(t *testing.T) {
	env, cleanup := SetupTestEnvironment(t)
	defer cleanup()

	accessToken, _ := createUserAndGetTokens(t, env, "test@mail.ru", "test")
	fileID := createFileWithStatus(t, env, "download.pdf", 1024, "application/pdf", accessToken, file_version.FileStatusReady)

	w := env.NewRequestWithAuth(t, "GET", "/api/v1/files/"+fileID.String()+"/versions/0/content", nil, accessToken)

	require.Equal(t, 200, w.Code)

	response := ParseJSONResponse(t, w)
	assert.Contains(t, response, "download_url")
	assert.Contains(t, response, "expires_in")
	assert.NotEmpty(t, response["download_url"])
	assert.Equal(t, "1h", response["expires_in"])
}

func TestFiles_GetDownloadURL_SpecificVersion(t *testing.T) {
	env, cleanup := SetupTestEnvironment(t)
	defer cleanup()

	accessToken, _ := createUserAndGetTokens(t, env, "test@mail.ru", "test")
	fileID := createFileWithStatus(t, env, "versioned.pdf", 1024, "application/pdf", accessToken, file_version.FileStatusReady)

	w := env.NewRequestWithAuth(t, "GET", "/api/v1/files/"+fileID.String()+"/versions/1/content", nil, accessToken)

	require.Equal(t, 200, w.Code)

	response := ParseJSONResponse(t, w)
	assert.Contains(t, response, "download_url")
	assert.NotEmpty(t, response["download_url"])
}

func TestFiles_GetDownloadURL_AccessDenied(t *testing.T) {
	env, cleanup := SetupTestEnvironment(t)
	defer cleanup()

	token1, _ := createUserAndGetTokens(t, env, "user1@mail.ru", "User 1")
	fileID := createFileWithStatus(t, env, "private.pdf", 1024, "application/pdf", token1, file_version.FileStatusReady)

	token2, _ := createUserAndGetTokens(t, env, "user2@mail.ru", "User 2")
	w := env.NewRequestWithAuth(t, "GET", "/api/v1/files/"+fileID.String()+"/versions/1/content", nil, token2)

	require.Equal(t, 403, w.Code)
}

// Тесты для проверки статусов файлов - используем createFileAndUpload для реальной загрузки

func TestFiles_CheckStatusProgression(t *testing.T) {
	env, cleanup := SetupTestEnvironment(t)
	defer cleanup()

	accessToken, _ := createUserAndGetTokens(t, env, "test@mail.ru", "test")

	body := map[string]interface{}{
		"name": "status_test.pdf",
		"size": 1024,
		"mime": "application/pdf",
	}
	w := env.NewJSONRequestWithAuth(t, "POST", "/api/v1/files", body, accessToken)
	require.Equal(t, 201, w.Code)

	response := ParseJSONResponse(t, w)
	fileIDStr := response["file_id"].(string)
	fileID, _ := uuid.Parse(fileIDStr)
	uploadURL := response["upload_url"].(string)

	assert.Equal(t, "processing", response["status"])

	err := uploadFileToS3(t, uploadURL, make([]byte, 1024))
	require.NoError(t, err)

	waitForFileStatus(t, env, fileID, accessToken, file_version.FileStatusUploaded)

	w = env.NewRequestWithAuth(t, "GET", "/api/v1/files/"+fileID.String(), nil, accessToken)
	response = ParseJSONResponse(t, w)
	assert.Equal(t, "uploaded", response["status"])

	waitForFileStatus(t, env, fileID, accessToken, file_version.FileStatusReady)

	w = env.NewRequestWithAuth(t, "GET", "/api/v1/files/"+fileID.String(), nil, accessToken)
	response = ParseJSONResponse(t, w)
	assert.Equal(t, "ready", response["status"])
}

func TestFiles_CannotDownloadProcessingFile(t *testing.T) {
	env, cleanup := SetupTestEnvironment(t)
	defer cleanup()

	accessToken, _ := createUserAndGetTokens(t, env, "test@mail.ru", "test")

	body := map[string]interface{}{
		"name": "processing.pdf",
		"size": 1024,
		"mime": "application/pdf",
	}
	w := env.NewJSONRequestWithAuth(t, "POST", "/api/v1/files", body, accessToken)
	require.Equal(t, 201, w.Code)

	response := ParseJSONResponse(t, w)
	fileID := response["file_id"].(string)

	w = env.NewRequestWithAuth(t, "GET", "/api/v1/files/"+fileID+"/versions/1/content", nil, accessToken)

	assert.NotEqual(t, 200, w.Code)
}
