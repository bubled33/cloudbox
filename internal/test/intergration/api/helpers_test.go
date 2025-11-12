package api_test

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/yourusername/cloud-file-storage/internal/domain/file_version"
)

func createUserAndLogin(t *testing.T, env *TestEnv, email string, displayName string) string {
	accessToken, _ := createUserAndGetTokens(t, env, email, displayName)
	return accessToken
}

func createUserAndGetTokens(t *testing.T, env *TestEnv, email string, displayName string) (string, string) {
	body := map[string]interface{}{
		"email":        email,
		"display_name": displayName,
	}
	w := env.NewJSONRequest(t, "POST", "/api/v1/magic-links", body)
	require.Equal(t, 200, w.Code)

	token := env.MailSender.GetTokenForEmail(email)
	require.NotEmpty(t, token, "Could not get token for email %s", email)

	url := fmt.Sprintf("/api/v1/magic-links/%s?token=%s", token, token)
	w = env.NewRequest(t, "GET", url, nil)
	require.Equal(t, 200, w.Code, "Failed to verify magic link: %s", w.Body.String())

	response := ParseJSONResponse(t, w)

	accessToken, ok := response["access_token"].(string)
	require.True(t, ok, "access_token not found in response")

	refreshToken, ok := response["refresh_token"].(string)
	require.True(t, ok, "refresh_token not found in response")

	return accessToken, refreshToken
}

func createFile(t *testing.T, env *TestEnv, name string, size uint64, mime string, accessToken string) uuid.UUID {
	body := map[string]interface{}{
		"name": name,
		"size": size,
		"mime": mime,
	}
	w := env.NewJSONRequestWithAuth(t, "POST", "/api/v1/files", body, accessToken)
	require.Equal(t, 201, w.Code)

	response := ParseJSONResponse(t, w)
	fileIDStr, ok := response["file_id"].(string)
	require.True(t, ok, "file_id not found in response")

	fileID, err := uuid.Parse(fileIDStr)
	require.NoError(t, err, "failed to parse file_id as UUID")
	return fileID
}

// createFileWithStatus создает файл и устанавливает нужный статус через domain методы и репозитории
func createFileWithStatus(t *testing.T, env *TestEnv, name string, size uint64, mime string, accessToken string, status file_version.FileStatus) uuid.UUID {
	body := map[string]interface{}{
		"name": name,
		"size": size,
		"mime": mime,
	}
	w := env.NewJSONRequestWithAuth(t, "POST", "/api/v1/files", body, accessToken)
	require.Equal(t, 201, w.Code, "Failed to create file: %s", w.Body.String())

	response := ParseJSONResponse(t, w)

	fileIDStr, ok := response["file_id"].(string)
	require.True(t, ok, "file_id not found or not a string")

	versionIDStr, ok := response["version_id"].(string)
	require.True(t, ok, "version_id not found")

	fileID, err := uuid.Parse(fileIDStr)
	require.NoError(t, err, "failed to parse file_id as UUID")

	versionID, err := uuid.Parse(versionIDStr)
	require.NoError(t, err, "failed to parse version_id as UUID")

	ctx := context.Background()

	// Используем UOW.Do для транзакции
	err = env.UOW.Do(ctx, func(ctx context.Context) error {
		fileEntity, err := env.FileService.GetByID(ctx, fileID)
		if err != nil {
			return fmt.Errorf("failed to get file: %w", err)
		}

		versionEntity, err := env.VersionService.GetVersionByID(ctx, versionID)
		if err != nil {
			return fmt.Errorf("failed to get file version: %w", err)
		}

		// Устанавливаем нужный статус через domain методы
		switch status {
		case file_version.FileStatusProcessing:
			fileEntity.MarkProcessing()
			versionEntity.MarkProcessing()

		case file_version.FileStatusUploaded:
			fileEntity.MarkUploaded()
			versionEntity.MarkUploaded()
			storageKeyStr := fmt.Sprintf("test-files/%s/%s", fileID.String(), versionID.String())
			storageKey, err := file_version.NewS3Key(storageKeyStr)
			if err != nil {
				return fmt.Errorf("failed to create S3Key: %w", err)
			}
			versionEntity.S3Key = storageKey

		case file_version.FileStatusReady:
			fileEntity.MarkReady()
			versionEntity.MarkReady()
			storageKeyStr := fmt.Sprintf("test-files/%s/%s", fileID.String(), versionID.String())
			storageKey, err := file_version.NewS3Key(storageKeyStr)
			if err != nil {
				return fmt.Errorf("failed to create storage S3Key: %w", err)
			}

			previewKeyStr := fmt.Sprintf("test-previews/%s/%s", fileID.String(), versionID.String())
			previewKey, err := file_version.NewS3Key(previewKeyStr)
			if err != nil {
				return fmt.Errorf("failed to create preview S3Key: %w", err)
			}

			versionEntity.S3Key = storageKey
			versionEntity.SetPreviewS3Key(previewKey)

		case file_version.FileStatusFailed:
			fileEntity.MarkFailed()
			versionEntity.MarkFailed()

		default:
			return fmt.Errorf("unknown status: %s", status)
		}

		// Сохраняем через репозитории внутри транзакции
		if err := env.FileCommandRepo.Save(ctx, fileEntity); err != nil {
			return fmt.Errorf("failed to save file: %w", err)
		}

		if err := env.FileVersionCommandRepo.Save(ctx, versionEntity); err != nil {
			return fmt.Errorf("failed to save file version: %w", err)
		}

		return nil
	})

	require.NoError(t, err, "Failed to update file status")

	return fileID
}

func createFileAndUpload(t *testing.T, env *TestEnv, name string, size uint64, mime string, accessToken string, status file_version.FileStatus) uuid.UUID {
	body := map[string]interface{}{
		"name": name,
		"size": size,
		"mime": mime,
	}
	w := env.NewJSONRequestWithAuth(t, "POST", "/api/v1/files", body, accessToken)
	require.Equal(t, 201, w.Code, "Failed to create file: %s", w.Body.String())

	response := ParseJSONResponse(t, w)

	fileIDStr, ok := response["file_id"].(string)
	require.True(t, ok, "file_id not found or not a string")

	versionIDStr, ok := response["version_id"].(string)
	require.True(t, ok, "version_id not found")

	uploadURL, ok := response["upload_url"].(string)
	require.True(t, ok, "upload_url not found")

	fileID, err := uuid.Parse(fileIDStr)
	require.NoError(t, err, "failed to parse file_id as UUID")

	err = uploadFileToS3(t, uploadURL, make([]byte, size))
	require.NoError(t, err, "failed to upload file to S3")

	waitForFileStatus(t, env, fileID, accessToken, status)
	_ = versionIDStr
	return fileID
}

func uploadFileToS3(t *testing.T, presignedURL string, data []byte) error {
	req, err := http.NewRequest("PUT", presignedURL, bytes.NewReader(data))
	if err != nil {
		return err
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("upload failed with status %d", resp.StatusCode)
	}

	return nil
}

func waitForFileStatus(t *testing.T, env *TestEnv, fileID uuid.UUID, accessToken string, req file_version.FileStatus, timeout ...time.Duration) {
	maxWait := 30 * time.Second
	if len(timeout) > 0 {
		maxWait = timeout[0]
	}

	deadline := time.Now().Add(maxWait)

	for time.Now().Before(deadline) {
		w := env.NewRequestWithAuth(t, "GET", "/api/v1/files/"+fileID.String(), nil, accessToken)
		if w.Code != 200 {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		response := ParseJSONResponse(t, w)
		status, ok := response["status"].(string)
		if ok && status == req.String() {
			return
		}

		time.Sleep(100 * time.Millisecond)
	}

	t.Fatalf("file %s did not reach status %s within %v", fileID, req.String(), maxWait)
}

func waitForFileStatusWithoutAuth(t *testing.T, env *TestEnv, fileID uuid.UUID, req file_version.FileStatus, timeout ...time.Duration) {
	maxWait := 30 * time.Second
	if len(timeout) > 0 {
		maxWait = timeout[0]
	}

	deadline := time.Now().Add(maxWait)

	for time.Now().Before(deadline) {
		var status string
		err := env.DB.DB.QueryRow(`SELECT status FROM files WHERE id = $1`, fileID).Scan(&status)

		if err == nil && status == req.String() {
			return
		}

		time.Sleep(100 * time.Millisecond)
	}

	t.Fatalf("file %s did not reach status %s within %v", fileID, req.String(), maxWait)
}

func getUserIDFromToken(t *testing.T, env *TestEnv, accessToken string) uuid.UUID {
	w := env.NewRequestWithAuth(t, "GET", "/api/v1/users/me", nil, accessToken)
	require.Equal(t, 200, w.Code)

	response := ParseJSONResponse(t, w)
	userIDStr, ok := response["id"].(string)
	require.True(t, ok, "id not found in response")

	userID, err := uuid.Parse(userIDStr)
	require.NoError(t, err, "failed to parse user_id as UUID")

	return userID
}

func createMultipleFiles(t *testing.T, env *TestEnv, count int, namePattern string, size uint64, mime string, accessToken string, status file_version.FileStatus) []uuid.UUID {
	fileIDs := make([]uuid.UUID, 0, count)

	for i := 0; i < count; i++ {
		name := fmt.Sprintf(namePattern, i)
		fileID := createFileWithStatus(t, env, name, size, mime, accessToken, status)
		fileIDs = append(fileIDs, fileID)
	}

	return fileIDs
}
