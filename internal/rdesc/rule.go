package rdesc

import "encoding/json"

type Act string

const (
	ActCreateProject Act = "create_project"
	ActDeleteProject Act = "delete_project"
	ActDoGlobalRules Act = "do_global_rules"
	ActQueryProject  Act = "query_project"
	ActChangeMode    Act = "change_mode"
	ActTest          Act = "test"
)

// Rule describes a rule to be run. Can be serialized and deserialized to/from JSON.
type Rule struct {
	// Name of the rule
	Act Act
	// Interval to run the rule. If not set, the rule will be run once.
	// Format:
	// - "random(5,10)" - run the rule randomly every 5-10 seconds
	// TODO: consider cron format
	Periodic string
	// Arguments passed to the rule constructor
	Args json.RawMessage
	// Timeout for rule execution.
	Timeout *Duration
	// Limit the execution frequency by setting minimal interval.
	MinInterval *Duration
}
