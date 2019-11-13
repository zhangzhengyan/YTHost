package Tool

import "errors"

func YTError(msg string) error  {
	return errors.New(msg)
}
