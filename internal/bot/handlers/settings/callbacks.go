package settings

import "strconv"

const (
	callbackListPrefix = "settings:group:"
)

func NewGroupSettingsData(groupID int64) string {
	return callbackListPrefix + strconv.FormatInt(groupID, 10)
}
