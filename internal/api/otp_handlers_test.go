package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestGenerateOTPCodeHandler(t *testing.T) {
	// Set Gin to test mode to reduce log output
	gin.SetMode(gin.TestMode)

	// Create a Gin router for test only
	router := gin.New()
	
	// Register the /v1/otp/code route
	v1 := router.Group("/v1")
	{
		otp := v1.Group("/otp")
		{
			otp.POST("/code", GenerateOTPCodeHandler)
		}
	}

	// Create a POST request with empty body
	req, err := http.NewRequest("POST", "/v1/otp/code", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Create a response recorder
	w := httptest.NewRecorder()

	// Perform the request
	router.ServeHTTP(w, req)

	// Assert HTTP status is 200
	assert.Equal(t, http.StatusOK, w.Code, "Expected HTTP status 200")

	// Parse the JSON response
	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	// Assert response contains 'code' field
	codeValue, exists := response["code"]
	assert.True(t, exists, "Response should contain 'code' field")

	// Assert code is a string
	code, ok := codeValue.(string)
	assert.True(t, ok, "Code should be a string")

	// Assert len(code) == 6
	assert.Equal(t, 6, len(code), "Code should be exactly 6 digits")

	// Assert all characters are digits
	matched, err := regexp.MatchString(`^\d{6}$`, code)
	if err != nil {
		t.Fatalf("Regex match error: %v", err)
	}
	assert.True(t, matched, "Code should contain only digits")
}

