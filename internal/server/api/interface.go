package api

type Service interface {
	ParseAndSave([]byte) error
	ParseAndGet([]byte) ([]byte, error)
	GetKnownMetrics() []string
	IsDBConnected() bool
	ParseAndSaveSeveral([]byte) error
}

type API interface {
	Run(string2 string) error
}
