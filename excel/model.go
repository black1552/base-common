package excel

type DropdownCol struct {
	Column   string   // 需要下拉框的列标识（如B, C）
	Options  []string // 下拉选项列表
	StartRow int      // 下拉框起始行（从1开始）
	EndRow   int      // 下拉框结束行
}
