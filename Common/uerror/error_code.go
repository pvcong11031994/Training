package uerror

const (
	// Error code and message
	// Error cdoe of user: 40000 < errorCode < 50000
	ErrorCodeUnanthorized = 40001
	MessageUnauthorized   = "Unauthorized"
	ErrorCodeForbidden    = 40002
	MessageForbidden      = "Forbidden"

	ErrorCodeBadRequest          = 40003
	MessageBadRequest            = "Bad request"
	ErrorCodeExists              = 40004
	MessageExists                = "Exists"
	ErrorCodeInternalServerError = 50000
	MessageInternalServerError   = "Internal server error"

	// HttpCode Unauthorized
	HttpCodeSuccess             = 200
	HttpCodeBadRequest          = 400
	HttpCodeUnanuthorized       = 401
	HttpCodeForbidden           = 403
	HttpCodeNotFound            = 404
	HttpCodeInternalServerError = 500
	HttpCodeOverLimitError      = 429
)
