package YTHostError

type YTError struct {
	Code int
	Msg string
}

func NewError(code int,msg string) *YTError {
	return &YTError{code,msg}
}

func (ye YTError)Error()string{
	return ye.Msg
}
