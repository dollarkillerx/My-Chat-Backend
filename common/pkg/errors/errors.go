package errors

import "fmt"

// 错误码定义
const (
	// 通用错误 1xxx
	ErrCodeSuccess       = 0
	ErrCodeUnknown       = 1000
	ErrCodeInvalidParam  = 1001
	ErrCodeUnauthorized  = 1002
	ErrCodeForbidden     = 1003
	ErrCodeNotFound      = 1004
	ErrCodeInternal      = 1005
	ErrCodeRateLimit     = 1006

	// 认证错误 2xxx
	ErrCodeInvalidToken  = 2001
	ErrCodeTokenExpired  = 2002
	ErrCodeLoginRequired = 2003

	// 用户错误 3xxx
	ErrCodeUserNotFound    = 3001
	ErrCodeUserExists      = 3002
	ErrCodePasswordWrong   = 3003
	ErrCodeUserDisabled    = 3004

	// 会话错误 4xxx
	ErrCodeConversationNotFound = 4001
	ErrCodeNotInConversation    = 4002
	ErrCodeConversationFull     = 4003

	// 消息错误 5xxx
	ErrCodeMessageNotFound   = 5001
	ErrCodeMessageTooLong    = 5002
	ErrCodeCannotRevoke      = 5003
	ErrCodeCannotEdit        = 5004
	ErrCodeRevokeTimeout     = 5005
	ErrCodeEditTimeout       = 5006

	// 关系错误 6xxx
	ErrCodeNotFriend        = 6001
	ErrCodeAlreadyFriend    = 6002
	ErrCodeBlocked          = 6003
	ErrCodeGroupNotFound    = 6004
	ErrCodeNotGroupMember   = 6005
	ErrCodeNoPermission     = 6006
)

// Error 业务错误
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// New 创建错误
func New(code int, message string) *Error {
	return &Error{Code: code, Message: message}
}

// Newf 创建格式化错误
func Newf(code int, format string, args ...interface{}) *Error {
	return &Error{Code: code, Message: fmt.Sprintf(format, args...)}
}

// 预定义错误
var (
	ErrSuccess       = New(ErrCodeSuccess, "success")
	ErrUnknown       = New(ErrCodeUnknown, "unknown error")
	ErrInvalidParam  = New(ErrCodeInvalidParam, "invalid parameter")
	ErrUnauthorized  = New(ErrCodeUnauthorized, "unauthorized")
	ErrForbidden     = New(ErrCodeForbidden, "forbidden")
	ErrNotFound      = New(ErrCodeNotFound, "not found")
	ErrInternal      = New(ErrCodeInternal, "internal error")
	ErrRateLimit     = New(ErrCodeRateLimit, "rate limit exceeded")

	ErrInvalidToken  = New(ErrCodeInvalidToken, "invalid token")
	ErrTokenExpired  = New(ErrCodeTokenExpired, "token expired")
	ErrLoginRequired = New(ErrCodeLoginRequired, "login required")

	ErrUserNotFound  = New(ErrCodeUserNotFound, "user not found")
	ErrUserExists    = New(ErrCodeUserExists, "user already exists")
	ErrPasswordWrong = New(ErrCodePasswordWrong, "wrong password")
	ErrUserDisabled  = New(ErrCodeUserDisabled, "user disabled")

	ErrConversationNotFound = New(ErrCodeConversationNotFound, "conversation not found")
	ErrNotInConversation    = New(ErrCodeNotInConversation, "not in conversation")
	ErrConversationFull     = New(ErrCodeConversationFull, "conversation is full")

	ErrMessageNotFound = New(ErrCodeMessageNotFound, "message not found")
	ErrMessageTooLong  = New(ErrCodeMessageTooLong, "message too long")
	ErrCannotRevoke    = New(ErrCodeCannotRevoke, "cannot revoke this message")
	ErrCannotEdit      = New(ErrCodeCannotEdit, "cannot edit this message")
	ErrRevokeTimeout   = New(ErrCodeRevokeTimeout, "revoke timeout exceeded")
	ErrEditTimeout     = New(ErrCodeEditTimeout, "edit timeout exceeded")

	ErrNotFriend      = New(ErrCodeNotFriend, "not friend")
	ErrAlreadyFriend  = New(ErrCodeAlreadyFriend, "already friend")
	ErrBlocked        = New(ErrCodeBlocked, "user blocked")
	ErrGroupNotFound  = New(ErrCodeGroupNotFound, "group not found")
	ErrNotGroupMember = New(ErrCodeNotGroupMember, "not group member")
	ErrNoPermission   = New(ErrCodeNoPermission, "no permission")
)

// IsError 判断是否为指定错误码
func IsError(err error, code int) bool {
	if e, ok := err.(*Error); ok {
		return e.Code == code
	}
	return false
}
