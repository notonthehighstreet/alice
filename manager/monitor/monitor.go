package monitor

type Monitor interface {
	GetUpdatedMetrics([]string) (*[]MetricUpdate, error)
}

type MetricUpdate struct {
	Name           string
	CurrentReading int
}
