package ledger

import (
	"encoding/json"
	"errors"
	"io"
)

func LoadBudgets(r io.Reader) error {
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()

	var items []Budget
	if err := dec.Decode(&items); err != nil {
		return errors.New("failed to parse budgets json: " + err.Error())
	}

	for _, b := range items {
		if _, err := SetBudget(b); err != nil {
			return errors.New("failed to set budget: " + err.Error())
		}
	}
	return nil
}
