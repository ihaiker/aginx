package ignore

type Ignore interface {
	//path必须为相对路径
	Add(path ...string)

	//path必须为相对路径
	Is(path string) bool

	//返回判断接口，如果返回false,将自动添加到
	IfNotIsAdd(path string) bool
}
