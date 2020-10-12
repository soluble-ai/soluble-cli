package policy

// Engine Policy Engine interface
type Engine interface {
	Init(string) error
	Configure() error
	Evaluate(EngineInput) (EngineOutput, error)
	GetResults() EngineOutput
	Release() error
}
