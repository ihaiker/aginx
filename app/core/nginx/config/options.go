package config

type Options struct {
	Delimiter        bool //是否允许分割符号。例如： server_name: nginx; 方式是否允许使用
	RemoveBrackets   bool //去除括号, 例如： name "aginx.io"; 这里的括号是否去除
	RemoveAnnotation bool //去除注解内容
}
