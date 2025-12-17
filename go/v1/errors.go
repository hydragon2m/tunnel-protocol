package v1

import "fmt"

/*
ErrorCode là mã lỗi ổn định, dùng cho:
- logging
- metrics
- hành vi (retry, close, ban...)
- tương lai: ERROR frame gửi qua wire
*/
type ErrorCode uint16

const (
	// ====== Generic ======
	ErrCodeUnknown ErrorCode = 0

	ErrCodeInvalidVersion ErrorCode = 1001
	ErrCodeFrameTooLarge  ErrorCode = 1002
	ErrCodeBadFrame       ErrorCode = 1003
	ErrCodeBadPayload     ErrorCode = 1004

	// ====== Auth / Handshake ======
	ErrCodeUnauthorized ErrorCode = 2001
	ErrCodeAuthExpired  ErrorCode = 2002

	// ====== Stream ======
	ErrCodeStreamNotFound ErrorCode = 3001
	ErrCodeStreamClosed  ErrorCode = 3002
)

/*
ProtocolError là error CHUẨN DUY NHẤT
được phép trả ra từ protocol layer
*/
type ProtocolError struct {
	Code ErrorCode
	Msg  string
}

func (e *ProtocolError) Error() string {
	if e.Msg == "" {
		return fmt.Sprintf("protocol error (%d)", e.Code)
	}
	return fmt.Sprintf("protocol error (%d): %s", e.Code, e.Msg)
}

/*
Helper tạo error nhanh, tránh copy-paste
*/
func NewError(code ErrorCode, msg string) *ProtocolError {
	return &ProtocolError{
		Code: code,
		Msg:  msg,
	}
}

/*
Helper check type an toàn
*/
func IsProtocolError(err error) (*ProtocolError, bool) {
	if err == nil {
		return nil, false
	}
	pe, ok := err.(*ProtocolError)
	return pe, ok
}
