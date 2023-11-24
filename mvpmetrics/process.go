package mvpmetrics

import "log"

type processMetrics struct {
	CPUTime    float64
	UserTime   float64
	KernelTime float64

	VirtualMemory  uint64
	ResidentMemory uint64

	OpenFiles int64

	Threads int64
	MinFlt  uint64
	MajFlt  uint64
}

func WriteProcessMetrics(mw *Writer) {
	var pm processMetrics
	err := collectProcessMetrics(&pm)
	if err != nil {
		log.Printf("ERROR: WriteProcessMetrics: %v", err)
		return
	}

	mw.WriteFloat("process_cpu_seconds_total", nil, nil, pm.CPUTime)
	mw.WriteFloat("process_cpu_user_seconds_total", nil, nil, pm.UserTime)
	mw.WriteUint("process_virtual_memory_bytes", nil, nil, pm.VirtualMemory)
	mw.WriteUint("process_resident_memory_bytes", nil, nil, pm.ResidentMemory)
	mw.WriteInt("process_open_fds", nil, nil, pm.OpenFiles)
	mw.WriteInt("process_threads", nil, nil, pm.Threads)
}
