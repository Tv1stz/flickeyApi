// Package sms provides a console SMS sender for development.
package sms

import (
	"context"
	"fmt"
	"log/slog"
)

// ConsoleSMSSender prints OTPs to stdout instead of sending real SMS.
// Used in dev environment.
type ConsoleSMSSender struct {
	Logger *slog.Logger
}

// NewConsoleSMSSender creates a new ConsoleSMSSender.
func NewConsoleSMSSender(logger *slog.Logger) *ConsoleSMSSender {
	return &ConsoleSMSSender{Logger: logger}
}

// SendOTP prints the OTP to stdout.
func (s *ConsoleSMSSender) SendOTP(_ context.Context, phone, otp string) error {
	msg := fmt.Sprintf("[SMS] → %s OTP: %s", phone, otp)
	fmt.Println(msg)
	s.Logger.Info("otp_sent_console", "phone_tail", phone[len(phone)-4:], "otp", otp)
	return nil
}
