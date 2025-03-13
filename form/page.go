package form

type ReqPageFunc interface {
	GetPage() int
	GetPerPage() int
}

// PageReq 分页请求
type PageReq struct {
	Page    int `form:"page" json:"page" example:"10"`
	PerPage int `form:"page_size" json:"page_size" example:"1"`
}

// GetPage 获取当前页码
func (req *PageReq) GetPage() int {
	return req.Page
}

// GetPerPage 获取每页数量
func (req *PageReq) GetPerPage() int {
	return req.PerPage
}

// Offset 计算分页偏移量
func (req *PageReq) GetOffset() int {
	return (req.Page - 1) * req.PerPage
}

// PageRes 分页响应
type PageRes struct {
	PageReq
	PageCount  int `json:"page_count" example:"0"`
	TotalCount int `json:"total_count" example:"0"`
}

// Pack 打包分页数据
func (res *PageRes) Pack(req ReqPageFunc, totalCount int) {
	res.TotalCount = totalCount
	res.PageCount = CalPageCount(totalCount, req.GetPerPage())
	res.Page = req.GetPage()
	res.PerPage = req.GetPerPage()
}

func CalPageCount(totalCount int, perPage int) int {
	if perPage <= 0 {
		return 0
	}
	return (int(totalCount) + perPage - 1) / perPage
}

// CalPage 计算分页偏移量
func CalPage(page, perPage int) (newPage, newPerPage int, offset int) {
	if page <= 0 {
		newPage = 1
	} else {
		newPage = page
	}
	if perPage <= 0 {
		newPerPage = 10
	} else {
		newPerPage = perPage
	}

	offset = (newPage - 1) * newPerPage
	return
}
