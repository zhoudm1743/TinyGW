package util

import (
	"github.com/jinzhu/copier"
	"go.uber.org/zap"
)

type toolUtil struct{}

var ToolUtil = toolUtil{}

// Copy
func (t toolUtil) Copy(toValue interface{}, fromValue interface{}) interface{} {
	if err := copier.Copy(toValue, fromValue); err != nil {
		zap.S().Panic("copy error", zap.Error(err))
		return nil
	}
	return toValue
}
