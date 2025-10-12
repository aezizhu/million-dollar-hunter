package metrics

type Recorder interface {
	IncCounter(name string, labels map[string]string)
	ObserveDuration(name string, seconds float64, labels map[string]string)
}

type Noop struct{}

func (Noop) IncCounter(string, map[string]string) {}

func (Noop) ObserveDuration(string, float64, map[string]string) {}
