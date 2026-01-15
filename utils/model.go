package utils

type Paginator[I any] struct {
	Items    []I
	Total    int
	IsSimple bool
}

type Paginate struct {
	Limit int `d:"20" json:"page_size" v:"max:50"`
	Page  int `d:"1" dc:"页码" json:"page_num"`
}
