package v2

type Paginator[I any] struct {
	Items    []I
	Total    int
	IsSimple bool
}

type Paginate struct {
	PageSize int `d:"20" json:"page_size" v:"max:50"`
	PageNum  int `d:"1" dc:"页码" json:"page_num"`
}
