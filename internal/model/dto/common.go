package dto

type PagerIO struct {
	Index int `form:"page"` // 頁碼
	Size  int `form:"size"` // 筆數
}

type AdminVO struct {
	IsDeleted bool   `json:"is_deleted"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	DeletedAt string `json:"deleted_at"`
}

type PagerVO struct {
	Index int `json:"page"`  // 頁碼
	Size  int `json:"size"`  // 筆數
	Pages int `json:"pages"` // 總頁數
	Total int `json:"total"` // 總筆數
}

type ListVO struct {
	List  interface{} `json:"list"`
	Pager PagerVO     `json:"pager"`
}
