package facade

import "errors"

// ErrNilRarityCalculator signals that a nil rarity calculator was provided
var ErrNilRarityCalculator = errors.New("nil rarity calculator")
