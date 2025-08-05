package collector

import (
	"sync"
)

// 预定义缓冲区大小
const (
	SmallBufferSize  = 512
	MediumBufferSize = 1024
	LargeBufferSize  = 4096
)

var (
	// 小缓冲区池（用于小数据包）
	smallBufferPool = sync.Pool{
		New: func() interface{} {
			buf := make([]byte, SmallBufferSize)
			return &buf
		},
	}

	// 中等缓冲区池（用于标准数据包）
	mediumBufferPool = sync.Pool{
		New: func() interface{} {
			buf := make([]byte, MediumBufferSize)
			return &buf
		},
	}

	// 大缓冲区池（用于大数据包）
	largeBufferPool = sync.Pool{
		New: func() interface{} {
			buf := make([]byte, LargeBufferSize)
			return &buf
		},
	}
)

// GetSmallBuffer 获取小型缓冲区
func GetSmallBuffer() []byte {
	return *smallBufferPool.Get().(*[]byte)
}

// GetMediumBuffer 获取中型缓冲区
func GetMediumBuffer() []byte {
	return *mediumBufferPool.Get().(*[]byte)
}

// GetLargeBuffer 获取大型缓冲区
func GetLargeBuffer() []byte {
	return *largeBufferPool.Get().(*[]byte)
}

// ReturnSmallBuffer 归还小型缓冲区
func ReturnSmallBuffer(buf []byte) {
	if cap(buf) >= SmallBufferSize {
		// 重置缓冲区，避免内存泄漏
		buf = buf[:SmallBufferSize]
		for i := range buf {
			buf[i] = 0
		}
		smallBufferPool.Put(&buf)
	}
}

// ReturnMediumBuffer 归还中型缓冲区
func ReturnMediumBuffer(buf []byte) {
	if cap(buf) >= MediumBufferSize {
		// 重置缓冲区，避免内存泄漏
		buf = buf[:MediumBufferSize]
		for i := range buf {
			buf[i] = 0
		}
		mediumBufferPool.Put(&buf)
	}
}

// ReturnLargeBuffer 归还大型缓冲区
func ReturnLargeBuffer(buf []byte) {
	if cap(buf) >= LargeBufferSize {
		// 重置缓冲区，避免内存泄漏
		buf = buf[:LargeBufferSize]
		for i := range buf {
			buf[i] = 0
		}
		largeBufferPool.Put(&buf)
	}
}

// GetBuffer 根据需要的大小获取适当的缓冲区
func GetBuffer(size int) []byte {
	if size <= SmallBufferSize {
		return GetSmallBuffer()
	} else if size <= MediumBufferSize {
		return GetMediumBuffer()
	} else {
		return GetLargeBuffer()
	}
}

// ReturnBuffer 根据缓冲区大小归还到对应的池
func ReturnBuffer(buf []byte) {
	capacity := cap(buf)
	if capacity <= SmallBufferSize {
		ReturnSmallBuffer(buf)
	} else if capacity <= MediumBufferSize {
		ReturnMediumBuffer(buf)
	} else if capacity <= LargeBufferSize {
		ReturnLargeBuffer(buf)
	}
	// 超过大缓冲区大小的不进行复用，让GC回收
}
