package models

import (
	"github.com/kamva/mgm/v3"
	"go.mongodb.org/mongo-driver/bson"
)

type Balance struct {
	mgm.DefaultModel `bson:",inline"`
	Balance          float64 `json:"balance" bson:"balance"`
	SourceName       string  `json:"source_name" bson:"source_name"`
}

func (balance *Balance) FetchOrCreate(sourceName string) {
	err := mgm.Coll(balance).First(bson.M{"source_name": sourceName}, balance)

	if err != nil && err.Error() == "mongo: no documents in result" {
		balance.Balance = 0
		balance.SourceName = sourceName
		_ = mgm.Coll(balance).Create(balance)
	}

}
