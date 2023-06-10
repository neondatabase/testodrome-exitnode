package rules

// Just some rule that can be executed periodically.
type ExecutableRule interface {
	Execute() error
}
