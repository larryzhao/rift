package runner

type RunnerKind string

const (
	RunnerKindHysteria2 RunnerKind = "hysteria2"
)

type Runner interface {
	Kind() RunnerKind
	Run() (int, error)
	Stop() error
}

var runners = map[RunnerKind]Runner{}

func Register(runner Runner) {
	runners[runner.Kind()] = runner
}

func Find(kind RunnerKind) (Runner, bool) {
	runner, ok := runners[kind]
	return runner, ok
}
