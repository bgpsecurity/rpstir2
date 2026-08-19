package model

type CpuPerformanceModel struct {
	CoresCount uint64 `json:"coresCount"` // 4
	Mhz        uint64 `json:"Mhz"`        //3201
}
type CpusPerformanceModel struct {
	CpuPerformanceModels []CpuPerformanceModel `json:"cpuPerformanceModels"`
}

type MemoryPerformanceModel struct {
	Memory     uint64 `json:"memory"`     // 8517005312
	FreeMemory uint64 `json:"freeMemory"` // 8517005312
}

type DiskPerformanceModel struct {
	TotalSpace uint64 `json:"totalSpace"` // the '/' space, in windows the "C:" space: 8517005312
}

type NetPerformanceModel struct {
	Mbps uint64 `json:"mbps"` // 16 Mbps, 64 Mbps
}

type TaskPerformanceModel struct {
	SnapshotHistoryCount     int64 `json:"snapshotHistoryCount"`
	DeltaHistoryCount        int64 `json:"deltaHistoryCount"`
	SnapshotHistoryOkCount   int64 `json:"snapshotHistoryOkCount"`
	DeltaHistoryOkCount      int64 `json:"deltaHistoryOkCount"`
	SnapshotHistoryFailCount int64 `json:"snapshotHistoryFailCount"`
	DeltaHistoryFailCount    int64 `json:"deltaHistoryFailCount"`

	SnapshotInProgressCount int64 `json:"snapshotInProgressCount"`
	DeltaInProgressCount    int64 `json:"deltaInProgressCount"`
	SnapshotOkCount         int64 `json:"snapshotOkCount"`
	DeltaOkCount            int64 `json:"deltaOkCount"`
	SnapshotFailCount       int64 `json:"snapshotFailCount"`
	DeltaFailCount          int64 `json:"deltaFailCount"`
}

func (c *TaskPerformanceModel) SetSelectedCount(isSnapshot bool, deltaCount int64) {
	if isSnapshot {
		c.SnapshotHistoryCount++
		c.SnapshotInProgressCount++
	} else {
		c.DeltaHistoryCount += deltaCount
		c.DeltaInProgressCount += deltaCount
	}
}

func (c *TaskPerformanceModel) SetResultCount(isResultOk bool, isSnapshot bool, deltaCount int64) {
	if isSnapshot {
		if isResultOk {
			c.SnapshotHistoryOkCount++
			c.SnapshotOkCount++
		} else {
			c.SnapshotHistoryFailCount++
			c.SnapshotFailCount++
		}
		c.SnapshotInProgressCount--
		if c.SnapshotInProgressCount < 0 {
			c.SnapshotInProgressCount = 0
		}
	} else {
		if isResultOk {
			c.DeltaHistoryOkCount += deltaCount
			c.DeltaOkCount += deltaCount
		} else {
			c.DeltaHistoryFailCount += deltaCount
			c.DeltaFailCount += deltaCount
		}
		c.DeltaInProgressCount -= deltaCount
		if c.DeltaInProgressCount < 0 {
			c.DeltaInProgressCount = 0
		}
	}
}

type PerformanceModel struct {
	TaskPerformanceModel   *TaskPerformanceModel   `json:"task"`
	CpusPerformanceModel   *CpusPerformanceModel   `json:"cpus"`
	MemoryPerformanceModel *MemoryPerformanceModel `json:"mem"`
	DiskPerformanceModel   *DiskPerformanceModel   `json:"disk"`
	NetPerformanceModel    *NetPerformanceModel    `json:"net"`
}

func (c *PerformanceModel) SupplementInit() {
	if c.TaskPerformanceModel == nil {
		c.TaskPerformanceModel = &TaskPerformanceModel{}
	}
	if c.CpusPerformanceModel == nil {
		c.CpusPerformanceModel = &CpusPerformanceModel{}
	}
	if c.MemoryPerformanceModel == nil {
		c.MemoryPerformanceModel = &MemoryPerformanceModel{}
	}
	if c.DiskPerformanceModel == nil {
		c.DiskPerformanceModel = &DiskPerformanceModel{}
	}
	if c.NetPerformanceModel == nil {
		c.NetPerformanceModel = &NetPerformanceModel{}
	}

}
