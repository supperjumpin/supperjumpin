package bot

import "strings"

const stampCustomIDPrefix = "stamp:"

func ParseStampCustomID(customID string) (roundID, jumpID, stampID string, ok bool) {
	if !strings.HasPrefix(customID, stampCustomIDPrefix) {
		return "", "", "", false
	}
	rest := strings.TrimPrefix(customID, stampCustomIDPrefix)
	parts := strings.Split(rest, ":")
	if len(parts) != 3 {
		return "", "", "", false
	}
	roundID, jumpID, stampID = parts[0], parts[1], parts[2]
	if roundID == "" || jumpID == "" || stampID == "" {
		return "", "", "", false
	}
	return roundID, jumpID, stampID, true
}
