package interfaces

type ApiError struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

var _ error = (*ApiError)(nil)

func (e *ApiError) Error() string {
	return e.Message
}
