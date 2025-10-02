package util_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

var cookies = make(map[string]*http.Cookie)

func Request(e *echo.Echo, method string, target string, token *string, bodyBytes []byte) ([]byte, int) {
	var body io.Reader
	if len(bodyBytes) > 0 {
		body = bytes.NewBuffer(bodyBytes)
	} else {
		body = nil
	}
	connectRequest := httptest.NewRequest(method, target, body)
	connectRequest.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)
	if token != nil {
		connectRequest.Header.Add(echo.HeaderAuthorization, fmt.Sprintf("Bearer %s", *token))
	}
	recorder := httptest.NewRecorder()
	for k, cookie := range cookies {
		if cookie.Expires.Before(time.Now()) {
			delete(cookies, k)
			continue
		}
		connectRequest.AddCookie(cookie)
	}
	e.ServeHTTP(recorder, connectRequest)
	for _, retCookie := range recorder.Result().Cookies() {
		if retCookie.Expires.Before(time.Now()) {
			retCookie.Expires = time.Now().Add(time.Duration(retCookie.MaxAge) * time.Second)
		}
		cookies[retCookie.Name] = retCookie
	}

	return recorder.Body.Bytes(), recorder.Code
}

// camelToSnake converts a camelCase string to snake_case
func camelToSnake(camel string) string {
	var snake strings.Builder
	for i, r := range camel {
		if i > 0 && 'A' <= r && r <= 'Z' {
			snake.WriteByte('_')
		}
		snake.WriteRune(r)
	}
	return strings.ToLower(snake.String())
}

// transformMapKeys recursively transforms all camelCase keys in a map to snake_case
func transformMapKeys(m map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	for k, v := range m {
		snakeKey := camelToSnake(k)

		// Recursively transform nested maps
		if nestedMap, ok := v.(map[string]interface{}); ok {
			result[snakeKey] = transformMapKeys(nestedMap)
		} else if nestedSlice, ok := v.([]interface{}); ok {
			// Handle arrays/slices that might contain maps
			transformedSlice := make([]interface{}, len(nestedSlice))
			for i, item := range nestedSlice {
				if itemMap, ok := item.(map[string]interface{}); ok {
					transformedSlice[i] = transformMapKeys(itemMap)
				} else {
					transformedSlice[i] = item
				}
			}
			result[snakeKey] = transformedSlice
		} else {
			result[snakeKey] = v
		}
	}

	return result
}

func RequestHTTP[T any](e *echo.Echo, method string, target string, token *string, body any) (T, int, error) {
	var res T
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		fmt.Printf("Error RequestHTTP: %v", err)
		return res, 0, err
	}
	resBytes, code := Request(e, method, target, token, bodyBytes)
	err = json.Unmarshal(resBytes, &res)

	return res, code, err
}
