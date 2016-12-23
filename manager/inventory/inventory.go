package inventory

type Inventory interface {
	Total() (int, error)
	Increase() error
	Decrease() error
	Status() Status
}

type Status int

const (
	OK Status = iota
	UPDATING
	FAILED
)
