package intents

type OpLogClose struct{}

func (OpLogClose) isIntent() {}

// Delta is positive for down and negative for up.
type OpLogNavigate struct {
	Delta int
}

func (OpLogNavigate) isIntent() {}

type OpLogShowDiff struct{}

func (OpLogShowDiff) isIntent() {}

type OpLogRestore struct{}

func (OpLogRestore) isIntent() {}

type OpLogRevert struct{}

func (OpLogRevert) isIntent() {}
