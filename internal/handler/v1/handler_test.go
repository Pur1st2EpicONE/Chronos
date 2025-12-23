package v1

import (
	"Chronos/internal/errs"
	"Chronos/internal/models"
	serviceMock "Chronos/internal/service/mocks"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	gomock "go.uber.org/mock/gomock"
)

func assertErrorResponse(t *testing.T, w *httptest.ResponseRecorder, code int, msg string) {
	assert.Equal(t, code, w.Code)
	var resp map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, msg, resp["error"])
}

func TestHandler_CreateNotification_Success(t *testing.T) {

	controller := gomock.NewController(t)
	defer controller.Finish()

	mockService := serviceMock.NewMockService(controller)
	handler := NewHandler(mockService)

	gin.SetMode(gin.TestMode)

	sendAt := time.Now().Add(time.Hour).UTC().Format(time.RFC3339)
	body, _ := json.Marshal(CreateNotificationV1{
		Channel: "email",
		Subject: "subject",
		Message: "qweqweqweqwe",
		SendAt:  sendAt,
		SendTo:  []string{"qwe@qweqweq.com"},
	})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	mockService.EXPECT().CreateNotification(gomock.Any(), gomock.Any()).Return("notif123", nil)

	handler.CreateNotification(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "notif123", resp["result"])

}

func TestHandler_CreateNotification_ErrInvalidJSON(t *testing.T) {

	controller := gomock.NewController(t)
	defer controller.Finish()

	mockService := serviceMock.NewMockService(controller)
	handler := NewHandler(mockService)

	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte("{invalid json}")))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.CreateNotification(c)

	assertErrorResponse(t, w, http.StatusBadRequest, errs.ErrInvalidJSON.Error())

}

func TestHandler_GetNotification_Success(t *testing.T) {

	controller := gomock.NewController(t)
	defer controller.Finish()

	mockService := serviceMock.NewMockService(controller)
	handler := NewHandler(mockService)

	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/?id=00000000-0000-0000-0000-000000000001", nil)

	mockService.EXPECT().GetStatus(gomock.Any(), "00000000-0000-0000-0000-000000000001").Return(models.StatusPending, nil)

	handler.GetNotification(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "pending", resp["result"])

}

func TestHandler_GetNotification_ErrInvalidID(t *testing.T) {

	controller := gomock.NewController(t)
	defer controller.Finish()

	mockService := serviceMock.NewMockService(controller)
	handler := NewHandler(mockService)

	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/?id=invalid", nil)

	handler.GetNotification(c)

	assertErrorResponse(t, w, http.StatusBadRequest, errs.ErrInvalidNotificationID.Error())

}

func TestHandler_CancelNotification_Success(t *testing.T) {

	controller := gomock.NewController(t)
	defer controller.Finish()

	mockService := serviceMock.NewMockService(controller)
	handler := NewHandler(mockService)

	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodPost, "/?id=00000000-0000-0000-0000-000000000001", nil)

	mockService.EXPECT().CancelNotification(gomock.Any(), "00000000-0000-0000-0000-000000000001").Return(nil)

	handler.CancelNotification(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, models.StatusCanceled, resp["result"])

}

func TestHandler_CancelNotification_ErrCannotCancel(t *testing.T) {

	controller := gomock.NewController(t)
	defer controller.Finish()

	mockService := serviceMock.NewMockService(controller)
	handler := NewHandler(mockService)

	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodPost, "/?id=00000000-0000-0000-0000-000000000001", nil)

	mockService.EXPECT().CancelNotification(gomock.Any(), "00000000-0000-0000-0000-000000000001").Return(errs.ErrCannotCancel)

	handler.CancelNotification(c)

	assertErrorResponse(t, w, http.StatusBadRequest, errs.ErrCannotCancel.Error())

}

func TestHandler_CreateNotification_ErrInvalidSendAt(t *testing.T) {

	controller := gomock.NewController(t)
	defer controller.Finish()

	mockService := serviceMock.NewMockService(controller)
	handler := NewHandler(mockService)
	gin.SetMode(gin.TestMode)

	body, _ := json.Marshal(CreateNotificationV1{
		Channel: "email",
		Subject: "subj",
		Message: "msg",
		SendAt:  "not-a-date",
		SendTo:  []string{"a@b.com"},
	})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.CreateNotification(c)

	assertErrorResponse(t, w, http.StatusBadRequest, errs.ErrInvalidSendAt.Error())

}

func TestHandler_CreateNotification_ErrService(t *testing.T) {

	controller := gomock.NewController(t)
	defer controller.Finish()

	mockService := serviceMock.NewMockService(controller)
	handler := NewHandler(mockService)
	gin.SetMode(gin.TestMode)

	sendAt := time.Now().Add(time.Hour).UTC().Format(time.RFC3339)
	body, _ := json.Marshal(CreateNotificationV1{
		Channel: "email",
		Subject: "subj",
		Message: "msg",
		SendAt:  sendAt,
		SendTo:  []string{"a@b.com"},
	})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	mockService.EXPECT().CreateNotification(gomock.Any(), gomock.Any()).Return("", errs.ErrInternal)

	handler.CreateNotification(c)

	assertErrorResponse(t, w, http.StatusInternalServerError, errs.ErrInternal.Error())

}

func TestHandler_GetNotification_ErrService(t *testing.T) {

	controller := gomock.NewController(t)
	defer controller.Finish()

	mockService := serviceMock.NewMockService(controller)
	handler := NewHandler(mockService)
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/?id=00000000-0000-0000-0000-000000000001", nil)

	mockService.EXPECT().GetStatus(gomock.Any(), "00000000-0000-0000-0000-000000000001").Return("", errs.ErrNotificationNotFound)

	handler.GetNotification(c)

	assertErrorResponse(t, w, http.StatusNotFound, errs.ErrNotificationNotFound.Error())

}

func TestHandler_CancelNotification_ErrInvalidID(t *testing.T) {

	controller := gomock.NewController(t)
	defer controller.Finish()

	mockService := serviceMock.NewMockService(controller)
	handler := NewHandler(mockService)
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodPost, "/?id=invalid-id", nil)

	handler.CancelNotification(c)

	assertErrorResponse(t, w, http.StatusBadRequest, errs.ErrInvalidNotificationID.Error())

}

func TestParseTime_MissingSendAt(t *testing.T) {
	_, err := parseTime("")
	assert.ErrorIs(t, err, errs.ErrMissingSendAt)
}

func TestMapErrorToStatus_UrgentDeliveryFailed(t *testing.T) {
	code, msg := mapErrorToStatus(errs.ErrUrgentDeliveryFailed)
	assert.Equal(t, http.StatusInternalServerError, code)
	assert.Equal(t, errs.ErrUrgentDeliveryFailed.Error(), msg)
}
