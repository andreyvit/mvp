package mvputil

func EnumChunks[T any](list []T, chunkSize int, f func(chunk []T)) {
	for {
		n := len(list)
		if n == 0 {
			break
		} else if n > chunkSize {
			f(list[:chunkSize])
			list = list[chunkSize:]
		} else {
			f(list)
			list = nil
		}
	}
}
