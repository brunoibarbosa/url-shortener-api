package response

type ErrorCodeStruct struct {
	InvalidRequest string
	NotFound       string
}

var ErrorCode = ErrorCodeStruct{
	InvalidRequest: "INVALID_REQUEST",
	NotFound:       "NOT_FOUND",
}
