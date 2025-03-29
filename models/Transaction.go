package models

import "github.com/kamva/mgm/v3"

type Transection struct {
	mgm.DefaultModel `bson:",inline"`
	ExpenseLocation  string  `json:"expense_location" bson:"expense_location" form:"expense_location"`
	Category         string  `json:"category" bson:"category" form:"category"`
	Amount           float64 `json:"amount" bson:"amount" form:"amount"`
	ExpenseType      string  `json:"expense_type" bson:"expense_type"`
	Description      string  `json:"description" bson:"description" form:"description"`
	SourceName       string  `json:"source_name" bson:"source_name"`
}


