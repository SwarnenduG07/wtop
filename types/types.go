package types

type ProcessInfo struct {
	PID        int32
	PPID       int32
	Name       string
	User       string
	Priority   int32
	Nice       int32
	CPUPercent float64
	Memory     uint64
	MemPercent float32
	VirtMem    uint64
	ResMem     uint64
	ShrMem     uint64
	Status     string
	Command    string
	Threads    int32
	CreateTime int64
}
