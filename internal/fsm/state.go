package fsm

type State int

const (
	StateSelectProductType State = iota
	StateInputProductID
	StateProductAction
	StateSelectArticle
)
