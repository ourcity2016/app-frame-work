package filters

import "app-frame-work/common"

type Filter interface {
	DoFilter(request *common.Request) (*common.Response, bool)
}

type Filters struct {
	Filters []Filter
}

// AddFilter 添加过滤器
func (fc *Filters) AddFilter(f Filter) {
	fc.Filters = append(fc.Filters, f)
}

func (fc *Filters) Execute(input *common.Request) (*common.Response, bool) {
	for _, f := range fc.Filters {
		output, canContinue := f.DoFilter(input)
		if !canContinue {
			return output, canContinue
		}
	}
	return common.OkWithNil(), true
}
