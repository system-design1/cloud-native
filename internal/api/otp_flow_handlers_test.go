package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go-backend-service/internal/middleware"
	"go-backend-service/internal/otp"
	apperrors "go-backend-service/pkg/errors"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeOTPFlowService struct {
	sendResp   *otp.SendResponse
	sendErr    error
	verifyResp *otp.VerifyResponse
	verifyErr  error
	sendReq    otp.SendRequest
	verifyReq  otp.VerifyRequest
}

func (s *fakeOTPFlowService) SendOTP(ctx context.Context, req otp.SendRequest) (*otp.SendResponse, error) {
	s.sendReq = req
	if s.sendErr != nil {
		return nil, s.sendErr
	}
	return s.sendResp, nil
}

func (s *fakeOTPFlowService) VerifyOTP(ctx context.Context, req otp.VerifyRequest) (*otp.VerifyResponse, error) {
	s.verifyReq = req
	if s.verifyErr != nil {
		return nil, s.verifyErr
	}
	return s.verifyResp, nil
}

func TestSendOTPHandlerSuccess(t *testing.T) {
	service := &fakeOTPFlowService{sendResp: &otp.SendResponse{
		RequestID: "request-1",
		ExpiredAt: time.Date(2026, 5, 9, 12, 0, 0, 0, time.UTC),
	}}
	router := newOTPFlowTestRouter()
	router.POST("/v1/otp/send", SendOTPHandler(service))

	w := performJSONRequest(router, "POST", "/v1/otp/send", `{"tenant_id":42,"phone":"+989121234567","token":"token","metadata":{"source":"test"}}`)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp otp.SendResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "request-1", resp.RequestID)
	assert.Equal(t, int64(42), service.sendReq.TenantID)
	assert.Equal(t, "+989121234567", service.sendReq.Phone)
	assert.Equal(t, "token", service.sendReq.Token)
	assert.Equal(t, "test", service.sendReq.Metadata["source"])
}

func TestSendOTPHandlerInvalidJSON(t *testing.T) {
	service := &fakeOTPFlowService{}
	router := newOTPFlowTestRouter()
	router.POST("/v1/otp/send", SendOTPHandler(service))

	w := performJSONRequest(router, "POST", "/v1/otp/send", `{invalid-json`)

	assertErrorResponse(t, w, http.StatusBadRequest)
}

func TestSendOTPHandlerMissingTenantID(t *testing.T) {
	service := &fakeOTPFlowService{}
	router := newOTPFlowTestRouter()
	router.POST("/v1/otp/send", SendOTPHandler(service))

	w := performJSONRequest(router, "POST", "/v1/otp/send", `{"phone":"+989121234567"}`)

	assertErrorResponse(t, w, http.StatusBadRequest)
}

func TestSendOTPHandlerEmptyPhone(t *testing.T) {
	service := &fakeOTPFlowService{}
	router := newOTPFlowTestRouter()
	router.POST("/v1/otp/send", SendOTPHandler(service))

	w := performJSONRequest(router, "POST", "/v1/otp/send", `{"tenant_id":42,"phone":" "}`)

	assertErrorResponse(t, w, http.StatusBadRequest)
}

func TestSendOTPHandlerTenantDisabled(t *testing.T) {
	service := &fakeOTPFlowService{sendErr: otp.ErrTenantDisabled}
	router := newOTPFlowTestRouter()
	router.POST("/v1/otp/send", SendOTPHandler(service))

	w := performJSONRequest(router, "POST", "/v1/otp/send", `{"tenant_id":42,"phone":"+989121234567"}`)

	assertErrorResponse(t, w, http.StatusForbidden)
}

func TestSendOTPHandlerOTPAlreadyActive(t *testing.T) {
	service := &fakeOTPFlowService{sendErr: otp.ErrOTPAlreadyActive}
	router := newOTPFlowTestRouter()
	router.POST("/v1/otp/send", SendOTPHandler(service))

	w := performJSONRequest(router, "POST", "/v1/otp/send", `{"tenant_id":42,"phone":"+989121234567"}`)

	assertErrorResponse(t, w, http.StatusTooManyRequests)
}

func TestSendOTPHandlerProviderFailure(t *testing.T) {
	service := &fakeOTPFlowService{sendErr: otp.ErrSMSProviderFailed}
	router := newOTPFlowTestRouter()
	router.POST("/v1/otp/send", SendOTPHandler(service))

	w := performJSONRequest(router, "POST", "/v1/otp/send", `{"tenant_id":42,"phone":"+989121234567"}`)

	assertErrorResponse(t, w, http.StatusBadGateway)
}

func TestSendOTPHandlerGenericError(t *testing.T) {
	service := &fakeOTPFlowService{sendErr: errors.New("boom")}
	router := newOTPFlowTestRouter()
	router.POST("/v1/otp/send", SendOTPHandler(service))

	w := performJSONRequest(router, "POST", "/v1/otp/send", `{"tenant_id":42,"phone":"+989121234567"}`)

	assertErrorResponse(t, w, http.StatusInternalServerError)
}

func TestVerifyOTPHandlerSuccess(t *testing.T) {
	service := &fakeOTPFlowService{verifyResp: &otp.VerifyResponse{Verified: true, RequestID: "request-2"}}
	router := newOTPFlowTestRouter()
	router.POST("/v1/otp/verify", VerifyOTPHandler(service))

	w := performJSONRequest(router, "POST", "/v1/otp/verify", `{"tenant_id":42,"phone":"+989121234567","code":"123456"}`)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp otp.VerifyResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp.Verified)
	assert.Equal(t, "request-2", resp.RequestID)
	assert.Equal(t, int64(42), service.verifyReq.TenantID)
	assert.Equal(t, "+989121234567", service.verifyReq.Phone)
	assert.Equal(t, "123456", service.verifyReq.Code)
}

func TestVerifyOTPHandlerBusinessFailure(t *testing.T) {
	service := &fakeOTPFlowService{verifyResp: &otp.VerifyResponse{
		Verified:  false,
		RequestID: "request-3",
		Reason:    otp.ReasonInvalidCode,
	}}
	router := newOTPFlowTestRouter()
	router.POST("/v1/otp/verify", VerifyOTPHandler(service))

	w := performJSONRequest(router, "POST", "/v1/otp/verify", `{"tenant_id":42,"phone":"+989121234567","code":"000000"}`)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp otp.VerifyResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.False(t, resp.Verified)
	assert.Equal(t, otp.ReasonInvalidCode, resp.Reason)
}

func TestVerifyOTPHandlerInvalidJSON(t *testing.T) {
	service := &fakeOTPFlowService{}
	router := newOTPFlowTestRouter()
	router.POST("/v1/otp/verify", VerifyOTPHandler(service))

	w := performJSONRequest(router, "POST", "/v1/otp/verify", `{invalid-json`)

	assertErrorResponse(t, w, http.StatusBadRequest)
}

func TestVerifyOTPHandlerMissingTenantID(t *testing.T) {
	service := &fakeOTPFlowService{}
	router := newOTPFlowTestRouter()
	router.POST("/v1/otp/verify", VerifyOTPHandler(service))

	w := performJSONRequest(router, "POST", "/v1/otp/verify", `{"phone":"+989121234567","code":"123456"}`)

	assertErrorResponse(t, w, http.StatusBadRequest)
}

func TestVerifyOTPHandlerEmptyPhone(t *testing.T) {
	service := &fakeOTPFlowService{}
	router := newOTPFlowTestRouter()
	router.POST("/v1/otp/verify", VerifyOTPHandler(service))

	w := performJSONRequest(router, "POST", "/v1/otp/verify", `{"tenant_id":42,"phone":" ","code":"123456"}`)

	assertErrorResponse(t, w, http.StatusBadRequest)
}

func TestVerifyOTPHandlerEmptyCode(t *testing.T) {
	service := &fakeOTPFlowService{}
	router := newOTPFlowTestRouter()
	router.POST("/v1/otp/verify", VerifyOTPHandler(service))

	w := performJSONRequest(router, "POST", "/v1/otp/verify", `{"tenant_id":42,"phone":"+989121234567","code":" "}`)

	assertErrorResponse(t, w, http.StatusBadRequest)
}

func TestVerifyOTPHandlerGenericError(t *testing.T) {
	service := &fakeOTPFlowService{verifyErr: errors.New("boom")}
	router := newOTPFlowTestRouter()
	router.POST("/v1/otp/verify", VerifyOTPHandler(service))

	w := performJSONRequest(router, "POST", "/v1/otp/verify", `{"tenant_id":42,"phone":"+989121234567","code":"123456"}`)

	assertErrorResponse(t, w, http.StatusInternalServerError)
}

func newOTPFlowTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.ErrorHandlerMiddleware())
	return router
}

func performJSONRequest(router *gin.Engine, method string, path string, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func assertErrorResponse(t *testing.T, w *httptest.ResponseRecorder, status int) {
	t.Helper()

	assert.Equal(t, status, w.Code)
	var resp apperrors.ErrorResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, status, resp.Code)
	assert.NotEmpty(t, resp.Message)
}
