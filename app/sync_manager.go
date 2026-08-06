package app

type SyncStatus int

const (
	SyncIdle SyncStatus = iota
	SyncRunning
	SyncFinished
)

type SyncManager struct {
	Status SyncStatus
}

var NodeSync = &SyncManager{
	Status: SyncIdle,
}

func (s *SyncManager) Start() {
	s.Status = SyncRunning
	LogInfo("Blockchain Sync Started")
}

func (s *SyncManager) Finish() {
	s.Status = SyncFinished
	LogInfo("Blockchain Sync Finished")
}

func (s *SyncManager) IsRunning() bool {
	return s.Status == SyncRunning
}

func (s *SyncManager) IsFinished() bool {
	return s.Status == SyncFinished
}
