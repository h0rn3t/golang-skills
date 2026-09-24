package probe

import "strconv"

// Label formats the id of name.
func Label(name string) string {
	userId := len(name)
	return strconv.Itoa(userId)
}
