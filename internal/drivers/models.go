package drivers

// Single query to the driver.
type SingleQuery struct {
	Query  string `json:"query"`
	Params []any  `json:"params"`
}
