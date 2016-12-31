package strategy

type Strategy interface {
	Evaluate() (*Recommendation, error)
}

type Recommendation int

const (
	SCALEDOWN Recommendation = iota - 1
	HOLD
	SCALEUP
)
