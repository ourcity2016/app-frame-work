package myfilter

import (
	"app-frame-work/common"
	roomcommon "app-frame-work/demo/common"
	"app-frame-work/logger"
)

var myLogger = logger.BuildMyLogger()

type LoginFilter struct{}

func (f *LoginFilter) DoFilter(request *common.Request) (*common.Response, bool) {
	users := roomcommon.CreatedUsesList
	_, exists := users.UserMap[request.SessionID]
	if !exists {
		myLogger.Warn("user not exists, value:%v", request)
		responseInfo := common.ERRORWithoutLogin("please login first...")
		responseInfo.Request = *request
		return responseInfo, false
	}
	return common.OkWithNil(), true
}
