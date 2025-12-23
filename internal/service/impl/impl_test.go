package impl

import (
	mockBroker "Chronos/internal/broker/mocks"
	mockCache "Chronos/internal/cache/mocks"
	"Chronos/internal/errs"
	mockLogger "Chronos/internal/logger/mocks"
	"Chronos/internal/models"
	mockStorage "Chronos/internal/repository/mocks"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestService_CancelNotification(t *testing.T) {

	ctx := context.Background()
	notificationID := "aboba123"

	controller := gomock.NewController(t)
	defer controller.Finish()

	mockLogger := mockLogger.NewMockLogger(controller)
	mockCache := mockCache.NewMockCache(controller)
	mockStorage := mockStorage.NewMockStorage(controller)

	svc := &Service{
		logger:  mockLogger,
		cache:   mockCache,
		storage: mockStorage,
	}

	t.Run("already canceled in cache", func(t *testing.T) {
		mockCache.EXPECT().GetStatus(ctx, notificationID).Return(models.StatusCanceled, nil)

		err := svc.CancelNotification(ctx, notificationID)
		require.ErrorIs(t, err, errs.ErrAlreadyCanceled)
	})

	t.Run("cannot cancel in cache", func(t *testing.T) {
		mockCache.EXPECT().GetStatus(ctx, notificationID).Return(models.StatusSent, nil)

		err := svc.CancelNotification(ctx, notificationID)
		require.ErrorIs(t, err, errs.ErrCannotCancel)
	})

	t.Run("not found in storage", func(t *testing.T) {
		mockCache.EXPECT().GetStatus(ctx, notificationID).Return("", errors.New("cache miss"))
		mockStorage.EXPECT().SetStatus(ctx, notificationID, models.StatusCanceled).Return(errs.ErrNotificationNotFound)

		err := svc.CancelNotification(ctx, notificationID)
		require.ErrorIs(t, err, errs.ErrNotificationNotFound)
	})

	t.Run("cannot cancel in storage but DB returns current status canceled", func(t *testing.T) {
		mockCache.EXPECT().GetStatus(ctx, notificationID).Return("", errors.New("cache miss"))
		mockStorage.EXPECT().SetStatus(ctx, notificationID, models.StatusCanceled).Return(errs.ErrCannotCancel)
		mockStorage.EXPECT().GetStatus(ctx, notificationID).Return(models.StatusCanceled, nil)
		mockCache.EXPECT().SetStatus(ctx, notificationID, models.StatusCanceled).Return(errors.New("cache error"))
		mockLogger.EXPECT().LogError("service — failed to set notification status in cache", gomock.Any(), "layer", "service.impl")

		err := svc.CancelNotification(ctx, notificationID)
		require.ErrorIs(t, err, errs.ErrAlreadyCanceled)
	})

	t.Run("cannot cancel in storage but DB returns another status", func(t *testing.T) {
		mockCache.EXPECT().GetStatus(ctx, notificationID).Return("", errors.New("cache miss"))
		mockStorage.EXPECT().SetStatus(ctx, notificationID, models.StatusCanceled).Return(errs.ErrCannotCancel)
		mockStorage.EXPECT().GetStatus(ctx, notificationID).Return(models.StatusSent, nil)
		mockCache.EXPECT().SetStatus(ctx, notificationID, models.StatusSent).Return(nil)

		err := svc.CancelNotification(ctx, notificationID)
		require.ErrorIs(t, err, errs.ErrCannotCancel)
	})

	t.Run("generic storage error", func(t *testing.T) {
		mockCache.EXPECT().GetStatus(ctx, notificationID).Return("", errors.New("cache miss"))
		mockStorage.EXPECT().SetStatus(ctx, notificationID, models.StatusCanceled).Return(errors.New("DB down"))
		mockLogger.EXPECT().LogError("service — failed to set notification status in DB", gomock.Any(), "layer", "service.impl")

		err := svc.CancelNotification(ctx, notificationID)
		require.Error(t, err)
		assert.EqualError(t, err, "DB down")
	})

	t.Run("successfully cancel notification", func(t *testing.T) {
		mockCache.EXPECT().GetStatus(ctx, notificationID).Return("", errors.New("cache miss"))
		mockStorage.EXPECT().SetStatus(ctx, notificationID, models.StatusCanceled).Return(nil)
		mockCache.EXPECT().SetStatus(ctx, notificationID, models.StatusCanceled).Return(nil)

		err := svc.CancelNotification(ctx, notificationID)
		require.NoError(t, err)
	})

	t.Run("cannot cancel in storage and GetStatus returns error", func(t *testing.T) {
		mockCache.EXPECT().GetStatus(ctx, notificationID).Return("", errors.New("cache miss"))
		mockStorage.EXPECT().SetStatus(ctx, notificationID, models.StatusCanceled).Return(errs.ErrCannotCancel)
		mockStorage.EXPECT().GetStatus(ctx, notificationID).Return("", errors.New("DB down"))
		mockLogger.EXPECT().LogError("service — failed to get notification status from DB", gomock.Any(), "layer", "service.impl")

		err := svc.CancelNotification(ctx, notificationID)
		require.Error(t, err)
		assert.EqualError(t, err, "DB down")
	})

	t.Run("successfully cancel notification but cache SetStatus fails", func(t *testing.T) {
		mockCache.EXPECT().GetStatus(ctx, notificationID).Return("", errors.New("cache miss"))
		mockStorage.EXPECT().SetStatus(ctx, notificationID, models.StatusCanceled).Return(nil)
		mockCache.EXPECT().SetStatus(ctx, notificationID, models.StatusCanceled).Return(errors.New("cache down"))
		mockLogger.EXPECT().LogError("service — failed to set notification status in cache", gomock.Any(), "layer", "service.impl")

		err := svc.CancelNotification(ctx, notificationID)
		require.NoError(t, err)
	})

}

func TestService_CreateNotification(t *testing.T) {

	ctx := context.Background()
	controller := gomock.NewController(t)
	defer controller.Finish()

	mockLogger := mockLogger.NewMockLogger(controller)
	mockStorage := mockStorage.NewMockStorage(controller)
	mockBroker := mockBroker.NewMockBroker(controller)

	svc := &Service{
		logger:  mockLogger,
		storage: mockStorage,
		broker:  mockBroker,
	}

	notification := models.Notification{
		Channel: "email",
		Message: "hello",
		Subject: "test",
		SendTo:  []string{"qwe@qweqwe.com"},
		SendAt:  time.Now().Add(time.Minute),
	}

	t.Run("validateCreate error", func(t *testing.T) {
		invalid := notification
		invalid.Channel = ""
		id, err := svc.CreateNotification(ctx, invalid)
		require.Empty(t, id)
		require.Error(t, err)
	})

	t.Run("storage.CreateNotification fails", func(t *testing.T) {
		mockStorage.EXPECT().CreateNotification(ctx, gomock.Any()).Return(errors.New("db down"))
		mockLogger.EXPECT().LogError("service — failed to create notification", gomock.Any(), "layer", "service.impl")

		id, err := svc.CreateNotification(ctx, notification)
		require.Empty(t, id)
		require.EqualError(t, err, "db down")
	})

	t.Run("broker produce fails and within recovery window, storage.DeleteNotification succeeds", func(t *testing.T) {
		mockStorage.EXPECT().CreateNotification(ctx, gomock.Any()).Return(nil)
		mockBroker.EXPECT().Produce(ctx, gomock.Any()).Return(errors.New("broker down"))
		mockLogger.EXPECT().LogError("service — failed to produce notification", gomock.Any(), "layer", "service.impl")
		mockStorage.EXPECT().DeleteNotification(ctx, gomock.Any()).Return(nil)

		id, err := svc.CreateNotification(ctx, notification)
		require.Empty(t, id)
		require.ErrorIs(t, err, errs.ErrUrgentDeliveryFailed)
	})

	t.Run("broker produce fails and DeleteNotification fails", func(t *testing.T) {
		mockStorage.EXPECT().CreateNotification(ctx, gomock.Any()).Return(nil)
		mockBroker.EXPECT().Produce(ctx, gomock.Any()).Return(errors.New("broker down"))
		mockLogger.EXPECT().LogError("service — failed to produce notification", gomock.Any(), "layer", "service.impl")
		mockStorage.EXPECT().DeleteNotification(ctx, gomock.Any()).Return(errors.New("db delete fail"))
		mockLogger.EXPECT().LogError("service — failed to delete notification from db", gomock.Any(), "layer", "service.impl")

		id, err := svc.CreateNotification(ctx, notification)
		require.Empty(t, id)
		require.ErrorIs(t, err, errs.ErrUrgentDeliveryFailed)
	})

	t.Run("broker produce fails but SendAt beyond recovery window", func(t *testing.T) {
		longFuture := notification
		longFuture.SendAt = time.Now().Add(2 * time.Hour)

		mockStorage.EXPECT().CreateNotification(ctx, gomock.Any()).Return(nil)
		mockBroker.EXPECT().Produce(ctx, gomock.Any()).Return(errors.New("broker down"))
		mockLogger.EXPECT().LogError("service — failed to produce notification", gomock.Any(), "layer", "service.impl")

		id, err := svc.CreateNotification(ctx, longFuture)
		require.NotEmpty(t, id)
		require.NoError(t, err)
	})

	t.Run("success", func(t *testing.T) {
		mockStorage.EXPECT().CreateNotification(ctx, gomock.Any()).Return(nil)
		mockBroker.EXPECT().Produce(ctx, gomock.Any()).Return(nil)

		id, err := svc.CreateNotification(ctx, notification)
		require.NotEmpty(t, id)
		require.NoError(t, err)
	})
}

func TestService_GetStatus(t *testing.T) {

	ctx := context.Background()
	notificationID := "aboba123"

	controller := gomock.NewController(t)
	defer controller.Finish()

	mockLogger := mockLogger.NewMockLogger(controller)
	mockCache := mockCache.NewMockCache(controller)
	mockStorage := mockStorage.NewMockStorage(controller)

	svc := &Service{
		logger:  mockLogger,
		cache:   mockCache,
		storage: mockStorage,
	}

	t.Run("status fetched from storage, cache set succeeds", func(t *testing.T) {
		mockCache.EXPECT().GetStatus(ctx, notificationID).Return("", errors.New("cache miss"))
		mockStorage.EXPECT().GetStatus(ctx, notificationID).Return("sent", nil)
		mockCache.EXPECT().SetStatus(ctx, notificationID, "sent").Return(nil)
		mockLogger.EXPECT().Debug("service — notification status fetched from DB", "notificationID", notificationID, "layer", "service.impl")

		status, err := svc.GetStatus(ctx, notificationID)
		require.NoError(t, err)
		require.Equal(t, "sent", status)
	})

	t.Run("status fetched from storage, cache set fails", func(t *testing.T) {
		mockCache.EXPECT().GetStatus(ctx, notificationID).Return("", errors.New("cache miss"))
		mockStorage.EXPECT().GetStatus(ctx, notificationID).Return("failed", nil)
		mockCache.EXPECT().SetStatus(ctx, notificationID, "failed").Return(errors.New("cache error"))
		mockLogger.EXPECT().LogError("service — failed to set notification status in cache", gomock.Any(), "notificationID", notificationID, "layer", "service.impl")
		mockLogger.EXPECT().Debug("service — notification status fetched from DB", "notificationID", notificationID, "layer", "service.impl")

		status, err := svc.GetStatus(ctx, notificationID)
		require.NoError(t, err)
		require.Equal(t, "failed", status)
	})

	t.Run("storage returns ErrNotificationNotFound", func(t *testing.T) {
		mockCache.EXPECT().GetStatus(ctx, notificationID).Return("", errors.New("cache miss"))
		mockStorage.EXPECT().GetStatus(ctx, notificationID).Return("", errs.ErrNotificationNotFound)
		mockLogger.EXPECT().Debug("service — notification status fetched from DB", "notificationID", notificationID, "layer", "service.impl")
		mockLogger.EXPECT().LogError("service — failed to get notification status from DB", errs.ErrNotificationNotFound, "notificationID", notificationID, "layer", "service.impl")

		status, err := svc.GetStatus(ctx, notificationID)
		require.ErrorIs(t, err, errs.ErrNotificationNotFound)
		require.Empty(t, status)
	})

	t.Run("storage returns generic error", func(t *testing.T) {
		mockCache.EXPECT().GetStatus(ctx, notificationID).Return("", errors.New("cache miss"))
		mockStorage.EXPECT().GetStatus(ctx, notificationID).Return("", errors.New("DB down"))
		mockLogger.EXPECT().LogError("service — failed to get notification status from DB", gomock.Any(), "notificationID", notificationID, "layer", "service.impl")

		status, err := svc.GetStatus(ctx, notificationID)
		require.Error(t, err)
		require.Empty(t, status)
	})

	t.Run("status fetched from cache", func(t *testing.T) {
		mockCache.EXPECT().GetStatus(ctx, notificationID).Return("pending", nil)
		mockLogger.EXPECT().Debug("service — notification status fetched from cache", "notificationID", notificationID, "layer", "service.impl")

		status, err := svc.GetStatus(ctx, notificationID)
		require.NoError(t, err)
		require.Equal(t, "pending", status)
	})

}

func TestNewService(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	mockLogger := mockLogger.NewMockLogger(controller)
	mockBroker := mockBroker.NewMockBroker(controller)
	mockCache := mockCache.NewMockCache(controller)
	mockStorage := mockStorage.NewMockStorage(controller)

	svc := NewService(mockLogger, mockBroker, mockCache, mockStorage)

	require.NotNil(t, svc)
	require.Equal(t, mockLogger, svc.logger)
	require.Equal(t, mockBroker, svc.broker)
	require.Equal(t, mockCache, svc.cache)
	require.Equal(t, mockStorage, svc.storage)
}

func TestValidateCreate(t *testing.T) {
	now := time.Now().UTC()
	validEmail := "qweqwe@eqweq.com"
	validNotification := models.Notification{
		Channel: models.Email,
		Message: "hello",
		SendAt:  now.Add(time.Hour),
		SendTo:  []string{validEmail},
		Subject: "qqqqq",
	}

	t.Run("valid notification", func(t *testing.T) {
		err := validateCreate(validNotification)
		require.NoError(t, err)
	})

	t.Run("missing channel", func(t *testing.T) {
		n := validNotification
		n.Channel = ""
		err := validateCreate(n)
		require.ErrorIs(t, err, errs.ErrMissingChannel)
	})

	t.Run("unsupported channel", func(t *testing.T) {
		n := validNotification
		n.Channel = "fax"
		err := validateCreate(n)
		require.ErrorIs(t, err, errs.ErrUnsupportedChannel)
	})

	t.Run("message too long", func(t *testing.T) {
		n := validNotification
		n.Message = strings.Repeat("a", models.MaxMessageLength+1)
		err := validateCreate(n)
		require.ErrorIs(t, err, errs.ErrMessageTooLong)
	})

	t.Run("missing sendAt", func(t *testing.T) {
		n := validNotification
		n.SendAt = time.Time{}
		err := validateCreate(n)
		require.ErrorIs(t, err, errs.ErrMissingSendAt)
	})

	t.Run("sendAt in past", func(t *testing.T) {
		n := validNotification
		n.SendAt = now.Add(-time.Hour)
		err := validateCreate(n)
		require.ErrorIs(t, err, errs.ErrSendAtInPast)
	})

	t.Run("sendAt too far in future", func(t *testing.T) {
		n := validNotification
		n.SendAt = now.AddDate(2, 0, 0)
		err := validateCreate(n)
		require.ErrorIs(t, err, errs.ErrSendAtTooFar)
	})

	t.Run("missing recipient", func(t *testing.T) {
		n := validNotification
		n.SendTo = []string{}
		err := validateCreate(n)
		require.ErrorIs(t, err, errs.ErrMissingSendTo)
	})

	t.Run("missing email subject", func(t *testing.T) {
		n := validNotification
		n.Subject = ""
		err := validateCreate(n)
		require.ErrorIs(t, err, errs.ErrMissingEmailSubject)
	})

	t.Run("email subject too long", func(t *testing.T) {
		n := validNotification
		n.Subject = strings.Repeat("a", models.MaxSubjectLength+1)
		err := validateCreate(n)
		require.ErrorIs(t, err, errs.ErrEmailSubjectTooLong)
	})

	t.Run("invalid recipient format", func(t *testing.T) {
		n := validNotification
		n.SendTo = []string{"invalid-email"}
		err := validateCreate(n)
		require.ErrorIs(t, err, errs.ErrInvalidEmailFormat)
	})

	t.Run("recipient too long", func(t *testing.T) {
		n := validNotification
		n.SendTo = []string{strings.Repeat("a", models.MaxEmailLength+1) + "@test.com"}
		err := validateCreate(n)
		require.ErrorIs(t, err, errs.ErrRecipientTooLong)
	})

	t.Run("empty recipient string", func(t *testing.T) {
		n := validNotification
		n.SendTo = []string{""}
		err := validateCreate(n)
		require.ErrorIs(t, err, errs.ErrInvalidEmailFormat)
	})

	t.Run("recipient missing domain dot", func(t *testing.T) {
		n := validNotification
		n.SendTo = []string{"user@invalid"}
		err := validateCreate(n)
		require.ErrorIs(t, err, errs.ErrInvalidEmailFormat)
	})

	t.Run("recipient missing @ symbol", func(t *testing.T) {
		n := validNotification
		n.SendTo = []string{"invalid.com"}
		err := validateCreate(n)
		require.ErrorIs(t, err, errs.ErrInvalidEmailFormat)
	})

}
