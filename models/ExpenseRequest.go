package models


type ExpenseRequest struct {
	ExpenseType     string  `json:"expense_type" form:"expense_type"` // inbound, outbound, transfer
	ExpenseLocation string  `json:"expense_location,omitempty" form:"expense_location,omitempty"`
	Category        string  `json:"category,omitempty" form:"category,omitempty"`
	Amount          float64 `json:"amount" form:"amount"`
	Description     string  `json:"description,omitempty" form:"description,omitempty"`
	BankName        string  `json:"bank_name,omitempty" form:"bank_name,omitempty"` // For inbound & transfer
	FromBank        string  `json:"from_bank,omitempty" form:"from_bank,omitempty"` // For transfer
	ToBank          string  `json:"to_bank,omitempty" form:"to_bank,omitempty"`     // For transfer
}
