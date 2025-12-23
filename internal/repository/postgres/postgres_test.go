package postgres_test

import (
	"Chronos/internal/config"
	"Chronos/internal/errs"
	"Chronos/internal/logger"
	"Chronos/internal/models"
	"Chronos/internal/repository/postgres"
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	wbf "github.com/wb-go/wbf/config"
	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/retry"
)

var testStorage *postgres.Storage

func TestMain(m *testing.M) {

	if err := wbf.New().LoadEnvFiles("../../../.env"); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	cfg := config.Storage{
		Host:     "localhost",
		Port:     "5434",
		Username: os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		DBName:   "chronos_test",
		SSLMode:  "disable",
		QueryRetryStrategy: config.Producer{
			Attempts: 3,
			Delay:    100 * time.Millisecond,
			Backoff:  1.5,
		},
	}

	logger, _ := logger.NewLogger(config.Logger{Debug: true})

	db, err := dbpg.New(fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.DBName, cfg.SSLMode), nil, &dbpg.Options{})
	if err != nil {
		logger.LogFatal("postgres_test — failed to connect to test DB", err, "layer", "repository.postgres_test")
	}

	testStorage = postgres.NewStorage(logger, cfg, db)

	exitCode := m.Run()
	testStorage.Close()
	os.Exit(exitCode)

}

func TestCreateNotification_Errors(t *testing.T) {

	ctx := context.Background()

	notification1 := models.Notification{
		ID:        "duplicate-uuid",
		Channel:   models.Email,
		Message:   "Hello",
		Status:    models.StatusPending,
		SendAt:    time.Now().Add(time.Hour),
		UpdatedAt: time.Now(),
		SendTo:    []string{"aboba@testing.com"},
	}

	if err := testStorage.CreateNotification(ctx, notification1); err != nil {
		t.Fatalf("CreateNotification failed unexpectedly: %v", err)
	}

	err := testStorage.CreateNotification(ctx, notification1)
	if err != nil {
		t.Logf("expected error captured: %v", err)
	} else {
		t.Fatalf("expected error for duplicate UUID, got nil")
	}

	notification2 := models.Notification{
		ID:        "uuid-recipients-error",
		Channel:   models.Email,
		Message:   "qwerty message",
		Status:    models.StatusPending,
		SendAt:    time.Now().Add(1 * time.Hour),
		UpdatedAt: time.Now(),
		SendTo:    []string{strings.Repeat("a", 300) + "@qwerty.com"},
	}

	err = testStorage.CreateNotification(ctx, notification2)
	if err != nil {
		t.Logf("expected recipients error captured: %v", err)
	} else {
		t.Fatalf("expected error for invalid recipient, got nil")
	}

}

func TestCleanup(t *testing.T) {

	ctx := context.Background()

	notifications := []models.Notification{
		{
			ID:        fmt.Sprintf("cleanup-%d", time.Now().UnixNano()),
			Channel:   models.Email,
			Message:   "To be canceled",
			Status:    models.StatusCanceled,
			SendAt:    time.Now().Add(-2 * time.Hour),
			UpdatedAt: time.Now().Add(-2 * time.Hour),
			SendTo:    []string{"test1@qwerty.com"},
		},
		{
			ID:        fmt.Sprintf("cleanup-%d", time.Now().UnixNano()+1),
			Channel:   models.Email,
			Message:   "To be sent",
			Status:    models.StatusSent,
			SendAt:    time.Now().Add(-2 * time.Hour),
			UpdatedAt: time.Now().Add(-2 * time.Hour),
			SendTo:    []string{"test2@qwerty.com"},
		},
	}

	for _, n := range notifications {
		if err := testStorage.CreateNotification(ctx, n); err != nil {
			t.Fatalf("failed to insert notification: %v", err)
		}
	}

	testStorage.Cleanup(ctx)
	t.Log("Cleanup executed successfully")

	for _, n := range notifications {
		var count int
		err := testStorage.DB().Master.QueryRowContext(ctx, "SELECT COUNT(*) FROM Notifications WHERE uuid=$1", n.ID).Scan(&count)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if count != 0 {
			t.Fatalf("notification %s was not deleted", n.ID)
		}
	}

}

func TestDeleteNotification(t *testing.T) {

	ctx := context.Background()

	notification := models.Notification{
		ID:        fmt.Sprintf("delete-%d", time.Now().UnixNano()),
		Channel:   models.Email,
		Message:   "To be deleted",
		Status:    models.StatusPending,
		SendAt:    time.Now().Add(1 * time.Hour),
		UpdatedAt: time.Now(),
		SendTo:    []string{"qwer@qweqweq.com"},
	}

	if err := testStorage.CreateNotification(ctx, notification); err != nil {
		t.Fatalf("CreateNotification failed: %v", err)
	}

	if err := testStorage.DeleteNotification(ctx, notification.ID); err != nil {
		t.Fatalf("DeleteNotification failed: %v", err)
	}

	var count int
	err := testStorage.DB().Master.QueryRowContext(ctx, "SELECT COUNT(*) FROM Notifications WHERE uuid=$1", notification.ID).Scan(&count)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected 0 notifications, got %d", count)
	}

	err = testStorage.DeleteNotification(ctx, "non-existent-id")
	if err != errs.ErrNotificationNotFound {
		t.Fatalf("expected ErrNotificationNotFound, got %v", err)
	}

}

func TestGetStatus(t *testing.T) {

	ctx := context.Background()

	notification := models.Notification{
		ID:        fmt.Sprintf("status-%d", time.Now().UnixNano()),
		Channel:   models.Email,
		Message:   "Check status",
		Status:    models.StatusPending,
		SendAt:    time.Now().Add(1 * time.Hour),
		UpdatedAt: time.Now(),
		SendTo:    []string{"qwe@qwe.com"},
	}

	if err := testStorage.CreateNotification(ctx, notification); err != nil {
		t.Fatalf("CreateNotification failed: %v", err)
	}

	status, err := testStorage.GetStatus(ctx, notification.ID)
	if err != nil {
		t.Fatalf("GetStatus failed: %v", err)
	}
	if status != notification.Status {
		t.Fatalf("expected status %s, got %s", notification.Status, status)
	}

	_, err = testStorage.GetStatus(ctx, "non-existent-id")
	if err != errs.ErrNotificationNotFound {
		t.Fatalf("expected ErrNotificationNotFound, got %v", err)
	}

}

func TestMarkLates(t *testing.T) {

	ctx := context.Background()

	notifications := []models.Notification{
		{
			ID:        fmt.Sprintf("late-%d", time.Now().UnixNano()),
			Channel:   models.Email,
			Message:   "Already late",
			Status:    models.StatusPending,
			SendAt:    time.Now().Add(-1 * time.Hour),
			UpdatedAt: time.Now(),
			SendTo:    []string{"qqqq@qqqqq.com"},
		},
		{
			ID:        fmt.Sprintf("late-%d", time.Now().UnixNano()+1),
			Channel:   models.Email,
			Message:   "Not late yet",
			Status:    models.StatusPending,
			SendAt:    time.Now().Add(1 * time.Hour),
			UpdatedAt: time.Now(),
			SendTo:    []string{"wew@wewew.com"},
		},
	}

	for _, n := range notifications {
		if err := testStorage.CreateNotification(ctx, n); err != nil {
			t.Fatalf("CreateNotification failed: %v", err)
		}
	}

	uuids, err := testStorage.MarkLates(ctx)
	if err != nil {
		t.Fatalf("MarkLates failed: %v", err)
	}

	if len(uuids) != 1 || uuids[0] != notifications[0].ID {
		t.Fatalf("expected 1 late UUID (%s), got %v", notifications[0].ID, uuids)
	}

	var status string
	err = testStorage.DB().Master.QueryRowContext(ctx, "SELECT status FROM Notifications WHERE uuid=$1", notifications[0].ID).Scan(&status)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	if status != models.StatusLate {
		t.Fatalf("expected status 'late', got %s", status)
	}

	uuids, err = testStorage.MarkLates(ctx)
	if err != nil {
		t.Fatalf("MarkLates failed: %v", err)
	}
	if len(uuids) != 0 {
		t.Fatalf("expected 0 uuids, got %v", uuids)
	}

	_, err = testStorage.DB().QueryWithRetry(ctx, retry.Strategy{
		Attempts: 1,
		Delay:    1,
		Backoff:  1,
	}, "SELECT * FROM NonExistentTable;")
	if err != nil {
		t.Logf("SQL error captured as expected: %v", err)
	} else {
		t.Fatalf("expected SQL error, got nil")
	}

}

func TestRecover(t *testing.T) {

	ctx := context.Background()

	notifications := []models.Notification{
		{
			ID:        "recover-1",
			Channel:   models.Email,
			Message:   "Pending notification",
			Status:    models.StatusPending,
			SendAt:    time.Now().Add(1 * time.Hour),
			UpdatedAt: time.Now(),
			SendTo:    []string{"user1@example.com"},
		},
		{
			ID:        "recover-2",
			Channel:   models.Email,
			Message:   "Late notification",
			Status:    models.StatusLate,
			SendAt:    time.Now().Add(-1 * time.Hour),
			UpdatedAt: time.Now(),
			SendTo:    []string{"user2@example.com"},
		},
	}

	for _, n := range notifications {
		if err := testStorage.CreateNotification(ctx, n); err != nil {
			t.Fatalf("CreateNotification failed: %v", err)
		}
	}

	testStorage.Config().RecoverLimit = len(notifications)

	result, err := testStorage.Recover(ctx)
	if err != nil {
		t.Fatalf("Recover failed: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("expected 2 notifications, got %d", len(result))
	}

	for _, n := range result {
		t.Logf("Recovered notification: %s, status: %s", n.ID, n.Status)
	}

}

func TestSetStatus(t *testing.T) {

	ctx := context.Background()

	n := models.Notification{
		ID:        "setstatus-test-1",
		Channel:   models.Email,
		Message:   "Test message",
		Status:    models.StatusPending,
		SendAt:    time.Now().Add(1 * time.Hour),
		UpdatedAt: time.Now(),
		SendTo:    []string{"user@example.com"},
	}

	if err := testStorage.CreateNotification(ctx, n); err != nil {
		t.Fatalf("failed to create notification: %v", err)
	}

	if err := testStorage.SetStatus(ctx, n.ID, models.StatusSent); err != nil {
		t.Fatalf("SetStatus failed: %v", err)
	}

	status, err := testStorage.GetStatus(ctx, n.ID)
	if err != nil {
		t.Fatalf("GetStatus failed: %v", err)
	}
	if status != models.StatusSent {
		t.Fatalf("expected status %s, got %s", models.StatusSent, status)
	}

	if err := testStorage.SetStatus(ctx, n.ID, models.StatusCanceled); err != errs.ErrCannotCancel {
		t.Fatalf("expected ErrCannotCancel, got %v", err)
	}

	if err := testStorage.SetStatus(ctx, "nonexistent-id", models.StatusSent); err != errs.ErrNotificationNotFound {
		t.Fatalf("expected ErrNotificationNotFound, got %v", err)
	}

}

func TestGetAllStatuses(t *testing.T) {

	ctx := context.Background()

	notifications := []models.Notification{
		{
			ID:        fmt.Sprintf("status-1-%d", time.Now().UnixNano()),
			Channel:   models.Email,
			Message:   "First notification",
			Status:    models.StatusPending,
			SendAt:    time.Now().Add(1 * time.Hour),
			UpdatedAt: time.Now(),
			SendTo:    []string{"first@qweqwe.com"},
		},
		{
			ID:        fmt.Sprintf("status-2-%d", time.Now().UnixNano()+1),
			Channel:   models.Email,
			Message:   "Second notification",
			Status:    models.StatusPending,
			SendAt:    time.Now().Add(2 * time.Hour),
			UpdatedAt: time.Now(),
			SendTo:    []string{"second@qweqwe.com"},
		},
	}

	for _, n := range notifications {
		if err := testStorage.CreateNotification(ctx, n); err != nil {
			t.Fatalf("CreateNotification failed: %v", err)
		}
	}

	allStatuses, err := testStorage.GetAllStatuses(ctx)
	if err != nil {
		t.Fatalf("GetAllStatuses failed: %v", err)
	}

	if len(allStatuses) < len(notifications) {
		t.Fatalf("expected at least %d notifications, got %d", len(notifications), len(allStatuses))
	}

	found := map[string]bool{}
	for _, n := range allStatuses {
		found[n.ID] = true
	}

	for _, n := range notifications {
		if !found[n.ID] {
			t.Fatalf("notification %s not found in GetAllStatuses result", n.ID)
		}
	}

	for i := 1; i < len(allStatuses); i++ {
		if allStatuses[i].SendAt.Before(allStatuses[i-1].SendAt) {
			t.Fatalf("notifications not sorted by SendAt")
		}
	}

}

func TestClose(t *testing.T) {
	log, _ := logger.NewLogger(config.Logger{Debug: true})
	db, _ := dbpg.New(fmt.Sprintf("host=localhost port=5434 user=%s password=%s dbname=chronos_test sslmode=disable",
		os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD")), nil, &dbpg.Options{})
	st := postgres.NewStorage(log, config.Storage{}, db)
	st.Close()
}
