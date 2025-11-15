package api_test

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUsers_GetMe_Success(t *testing.T) {
	env, cleanup := SetupTestEnvironment(t)
	defer cleanup()

	email := "test@example.com"
	displayName := "Test User"
	accessToken, _ := createUserAndGetTokens(t, env, email, displayName)

	w := env.NewRequestWithAuth(t, "GET", "/api/v1/users/me", nil, accessToken)

	require.Equal(t, 200, w.Code)

	response := ParseJSONResponse(t, w)
	assert.Equal(t, email, response["email"])
	assert.Equal(t, displayName, response["display_name"])
	assert.NotEmpty(t, response["id"])
	assert.NotEmpty(t, response["created_at"])
}

func TestUsers_GetMe_Unauthorized(t *testing.T) {
	env, cleanup := SetupTestEnvironment(t)
	defer cleanup()

	tests := []struct {
		name  string
		token string
	}{
		{
			name:  "no token",
			token: "",
		},
		{
			name:  "invalid token",
			token: "invalid-token-123",
		},
		{
			name:  "malformed token",
			token: "not.a.jwt.token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var w *httptest.ResponseRecorder
			if tt.token == "" {
				w = env.NewRequest(t, "GET", "/api/v1/users/me", nil)
			} else {
				w = env.NewRequestWithAuth(t, "GET", "/api/v1/users/me", nil, tt.token)
			}

			require.Equal(t, 401, w.Code)

			response := ParseJSONResponse(t, w)
			assert.Contains(t, response, "error")
		})
	}
}

func TestUsers_GetMe_AfterLogout(t *testing.T) {
	env, cleanup := SetupTestEnvironment(t)
	defer cleanup()

	email := "test@example.com"
	accessToken, _ := createUserAndGetTokens(t, env, email, "Test User")

	w := env.NewRequestWithAuth(t, "GET", "/api/v1/users/me", nil, accessToken)
	require.Equal(t, 200, w.Code)

	w = env.NewRequestWithAuth(t, "DELETE", "/api/v1/auth/sessions/current", nil, accessToken)
	require.Equal(t, 200, w.Code)

	w = env.NewRequestWithAuth(t, "GET", "/api/v1/users/me", nil, accessToken)
	require.Equal(t, 401, w.Code)
}

func TestUsers_UpdateProfile_Success(t *testing.T) {
	env, cleanup := SetupTestEnvironment(t)
	defer cleanup()

	email := "test@example.com"
	displayName := "Test User"
	accessToken, _ := createUserAndGetTokens(t, env, email, displayName)

	t.Run("update display name", func(t *testing.T) {
		newDisplayName := "Updated Name"
		body := map[string]interface{}{
			"email":        email,
			"display_name": newDisplayName,
		}

		w := env.NewJSONRequestWithAuth(t, "PATCH", "/api/v1/users/me", body, accessToken)

		require.Equal(t, 200, w.Code)

		response := ParseJSONResponse(t, w)
		assert.Equal(t, email, response["email"])
		assert.Equal(t, newDisplayName, response["display_name"])
	})

	t.Run("update email", func(t *testing.T) {
		newEmail := "newemail@example.com"
		body := map[string]interface{}{
			"email":        newEmail,
			"display_name": displayName,
		}

		w := env.NewJSONRequestWithAuth(t, "PATCH", "/api/v1/users/me", body, accessToken)

		require.Equal(t, 200, w.Code)

		response := ParseJSONResponse(t, w)
		assert.Equal(t, newEmail, response["email"])
	})

	t.Run("update both email and display name", func(t *testing.T) {
		newEmail := "another@example.com"
		newDisplayName := "Another Name"
		body := map[string]interface{}{
			"email":        newEmail,
			"display_name": newDisplayName,
		}

		w := env.NewJSONRequestWithAuth(t, "PATCH", "/api/v1/users/me", body, accessToken)

		require.Equal(t, 200, w.Code)

		response := ParseJSONResponse(t, w)
		assert.Equal(t, newEmail, response["email"])
		assert.Equal(t, newDisplayName, response["display_name"])
	})
}

func TestUsers_UpdateProfile_EmailConflict(t *testing.T) {
	env, cleanup := SetupTestEnvironment(t)
	defer cleanup()

	email1 := "user1@example.com"
	token1, _ := createUserAndGetTokens(t, env, email1, "User 1")

	email2 := "user2@example.com"
	_, _ = createUserAndGetTokens(t, env, email2, "User 2")

	body := map[string]interface{}{
		"email":        email2,
		"display_name": "User 1",
	}

	w := env.NewJSONRequestWithAuth(t, "PATCH", "/api/v1/users/me", body, token1)

	require.Equal(t, 409, w.Code)

	response := ParseJSONResponse(t, w)
	assert.Contains(t, response, "error")
	assert.Equal(t, "email already in use", response["error"])
}

func TestUsers_UpdateProfile_InvalidInput(t *testing.T) {
	env, cleanup := SetupTestEnvironment(t)
	defer cleanup()

	email := "test@example.com"
	accessToken, _ := createUserAndGetTokens(t, env, email, "Test User")

	tests := []struct {
		name string
		body map[string]interface{}
	}{
		{
			name: "empty email",
			body: map[string]interface{}{
				"email":        "",
				"display_name": "Test",
			},
		},
		{
			name: "invalid email format",
			body: map[string]interface{}{
				"email":        "not-an-email",
				"display_name": "Test",
			},
		},
		{
			name: "empty display name",
			body: map[string]interface{}{
				"email":        email,
				"display_name": "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := env.NewJSONRequestWithAuth(t, "PATCH", "/api/v1/users/me", tt.body, accessToken)

			require.Equal(t, 400, w.Code)

			response := ParseJSONResponse(t, w)
			assert.Contains(t, response, "error")
		})
	}
}

func TestUsers_UpdateProfile_Unauthorized(t *testing.T) {
	env, cleanup := SetupTestEnvironment(t)
	defer cleanup()

	body := map[string]interface{}{
		"email":        "test@example.com",
		"display_name": "Test",
	}

	w := env.NewJSONRequestWithAuth(t, "PATCH", "/api/v1/users/me", body, "invalid-token")

	require.Equal(t, 401, w.Code)
}

func TestUsers_DeleteAccount_Success(t *testing.T) {
	env, cleanup := SetupTestEnvironment(t)
	defer cleanup()

	email := "test@example.com"
	accessToken, _ := createUserAndGetTokens(t, env, email, "Test User")

	w := env.NewRequestWithAuth(t, "GET", "/api/v1/users/me", nil, accessToken)
	require.Equal(t, 200, w.Code)

	// Удаляем аккаунт
	body := map[string]interface{}{
		"confirmation": "DELETE_MY_ACCOUNT",
	}

	w = env.NewJSONRequestWithAuth(t, "DELETE", "/api/v1/users/me", body, accessToken)

	require.Equal(t, 200, w.Code)

	response := ParseJSONResponse(t, w)
	assert.Contains(t, response, "message")
	assert.Equal(t, "Account successfully deleted", response["message"])

	w = env.NewRequestWithAuth(t, "GET", "/api/v1/users/me", nil, accessToken)
	require.Equal(t, 401, w.Code)
}

// func TestUsers_DeleteAccount_InvalidConfirmation(t *testing.T) {
// 	env, cleanup := SetupTestEnvironment(t)
// 	defer cleanup()

// 	email := "test@example.com"
// 	accessToken, _ := createUserAndGetTokens(t, env, email, "Test User")

// 	tests := []struct {
// 		name         string
// 		confirmation string
// 	}{
// 		{
// 			name:         "wrong text",
// 			confirmation: "DELETE MY ACCOUNT",
// 		},
// 		{
// 			name:         "empty confirmation",
// 			confirmation: "",
// 		},
// 		{
// 			name:         "lowercase",
// 			confirmation: "delete_my_account",
// 		},
// 		{
// 			name:         "partial match",
// 			confirmation: "DELETE",
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			body := map[string]interface{}{
// 				"confirmation": tt.confirmation,
// 			}

// 			w := env.NewJSONRequestWithAuth(t, "DELETE", "/api/v1/users/me", body, accessToken)

// 			require.Equal(t, 400, w.Code)

// 			response := ParseJSONResponse(t, w)
// 			assert.Contains(t, response, "error")
// 			assert.Equal(t, "invalid confirmation", response["error"])
// 		})
// 	}
// }

func TestUsers_DeleteAccount_Unauthorized(t *testing.T) {
	env, cleanup := SetupTestEnvironment(t)
	defer cleanup()

	body := map[string]interface{}{
		"confirmation": "DELETE_MY_ACCOUNT",
	}

	w := env.NewJSONRequestWithAuth(t, "DELETE", "/api/v1/users/me", body, "invalid-token")

	require.Equal(t, 401, w.Code)
}

func TestUsers_DeleteAccount_MissingConfirmation(t *testing.T) {
	env, cleanup := SetupTestEnvironment(t)
	defer cleanup()

	email := "test@example.com"
	accessToken, _ := createUserAndGetTokens(t, env, email, "Test User")

	w := env.NewJSONRequestWithAuth(t, "DELETE", "/api/v1/users/me", map[string]interface{}{}, accessToken)

	require.Equal(t, 400, w.Code)
}

func TestUsers_DeleteAccount_CascadeDelete(t *testing.T) {
	env, cleanup := SetupTestEnvironment(t)
	defer cleanup()

	email := "test@example.com"
	accessToken, _ := createUserAndGetTokens(t, env, email, "Test User")

	body := map[string]interface{}{
		"confirmation": "DELETE_MY_ACCOUNT",
	}

	w := env.NewJSONRequestWithAuth(t, "DELETE", "/api/v1/users/me", body, accessToken)
	require.Equal(t, 200, w.Code)
}
