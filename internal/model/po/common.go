package po

import "time"

const (
	defaultPageIndex = 1
	defaultPageSize  = 10
)

type BaseTimeColumns struct {
	CreatedAt time.Time `gorm:"created_at;autoCreateTime"`
	UpdatedAt time.Time `gorm:"updated_at;autoUpdateTime"`
}

type DeleteRelatedColumns struct {
	IsDeleted bool       `gorm:"is_deleted"`
	DeletedAt *time.Time `gorm:"deleted_at"`
}

type Pager struct {
	Index int    // 頁碼
	Size  int    // 筆數
	Order string // 排序
}

func (p *Pager) GetSize() int {
	if p.Size < 1 {
		return defaultPageSize
	}

	return p.Size
}

func (p *Pager) GetIndex() int {
	if p.Index < 1 {
		return defaultPageIndex
	}

	return p.Index
}

func (p *Pager) GetOffset() int {
	return p.GetSize() * (p.GetIndex() - 1)
}

func NewPagerResult(paging *Pager, total int64) *PagerResult {
	intTotal := int(total)
	totalPage := intTotal / paging.GetSize()
	if intTotal%paging.GetSize() > 0 {
		totalPage++
	}

	return &PagerResult{
		Index: paging.GetIndex(),
		Size:  paging.GetSize(),
		Pages: totalPage,
		Total: intTotal,
	}
}

type PagerResult struct {
	Index int
	Size  int
	Pages int
	Total int
}
