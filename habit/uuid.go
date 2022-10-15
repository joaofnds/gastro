package habit

import "regexp"

var uuidRegexp = regexp.MustCompile("[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}")

func IsUUID(uuid string) bool {
	if len(uuid) != 36 {
		return false
	}

	return uuidRegexp.Match([]byte(uuid))
}
