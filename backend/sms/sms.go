// Package sms provides the SMS sender abstraction and implementations.
package sms

import "context"

// SMSSender is the interface for sending SMS messages.
type SMSSender interface {
	SendOTP(ctx context.Context, phone, otp string) error
}
