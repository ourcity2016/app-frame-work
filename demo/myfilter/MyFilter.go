package myfilter

import "app-frame-work/common"

type RoomFilter struct{}

func (f *RoomFilter) DoFilter(request *common.Request) (*common.Response, bool) {
	return common.ERROR("进来Filter", "哈哈哈"), true
}
