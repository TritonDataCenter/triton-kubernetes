package testutils

type StepAction uint

const (
	Continue StepAction = iota
	Halt
)

const (
	StateCancelled = "cancelled"
	StateHalted    = "halted"
)

type Step interface {
	Run(TritonStateBag) StepAction

	Cleanup(TritonStateBag)
}
