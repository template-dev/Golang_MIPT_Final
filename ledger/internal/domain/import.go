package domain

type ImportItem struct {
	Index int
	Tx    Transaction
}

type ImportError struct {
	Index int    `json:"index"`
	Error string `json:"error"`
}

type ImportSummary struct {
	Accepted int64         `json:"accepted"`
	Rejected int64         `json:"rejected"`
	Errors   []ImportError `json:"errors"`
}
