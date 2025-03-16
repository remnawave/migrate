package remnawave

type ApiError struct {
	Timestamp string `json:"timestamp"`
	Path      string `json:"path"`
	Message   string `json:"message"`
	ErrorCode string `json:"errorCode"`
}

type UserExistsError struct {
	Username string
	ApiError ApiError
}

func (e *UserExistsError) Error() string {
	return e.ApiError.Message
}

func IsUserExistsError(err error) bool {
	if _, ok := err.(*UserExistsError); ok {
		return true
	}
	return false
}
