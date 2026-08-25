package sms_test

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"flickey/go-backend/sms"
)

func TestConsoleSMSSender_SendOTP(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	sender := sms.NewConsoleSMSSender(logger)

	ctx := context.Background()
	if err := sender.SendOTP(ctx, "+375291234567", "123456"); err != nil {
		t.Fatalf("SendOTP failed: %v", err)
	}
}
