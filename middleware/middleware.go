package middleware

// Middleware 数据帧处理中间件，当数据到达，或发送时数据将按中间件注册顺序进行处理
type Middleware interface {
	Write([]byte) []byte
	Read([]byte) []byte
}

type MiddlewareMngr struct {
	middlewares []Middleware
}

func New() *MiddlewareMngr {
	return &MiddlewareMngr{}
}

// 添加中间件
func (mwm *MiddlewareMngr) Use(mw Middleware) {
	mwm.middlewares = append(mwm.middlewares, mw)
}

func (mwm *MiddlewareMngr) Write(data []byte) []byte {
	return mwm.write(data, 0)
}

func (mwm *MiddlewareMngr) write(data []byte, i int) []byte {
	if i >= len(mwm.middlewares) {
		return data
	}
	return mwm.write(data, i+1)
}

func (mwm *MiddlewareMngr) Read(data []byte) []byte {
	return mwm.read(data, 0)
}
func (mwm *MiddlewareMngr) read(data []byte, i int) []byte {
	if i >= len(mwm.middlewares) {
		return data
	}
	return mwm.read(data, i+1)
}
