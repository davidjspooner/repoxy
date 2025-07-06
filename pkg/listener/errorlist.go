package listener

import "fmt"

type ErrorList []error

func (el *ErrorList) Add(err error) {
	if err == nil {
		return
	}
	*el = append(*el, err)
}
func (el ErrorList) Error() string {
	if len(el) == 0 {
		return ""
	}
	msg := "Multiple errors occurred:\n"
	for i, err := range el {
		msg += fmt.Sprintf("%d: %v\n", i+1, err)
	}
	return msg
}
func (el *ErrorList) IsEmpty() bool {
	if el == nil {
		return true
	}
	return len(*el) == 0
}
func (el *ErrorList) Clear() {
	if el == nil {
		return
	}
	*el = (*el)[:0]
}
