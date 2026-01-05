package settings

import "strconv"

func NewGroupSettingsData(groupID int64) string {
	return callbackGroupPrefix + strconv.FormatInt(groupID, 10)
}
