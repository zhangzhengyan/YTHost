package middleware

// Middleware 数据帧处理中间件，当数据到达，或发送时数据将按中间件注册顺序进行处理
type Middleware interface {
	OnDataFrame([]byte)[]byte
	SendDataFrame([]byte)[]byte
}
