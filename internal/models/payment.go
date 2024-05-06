package models

import (
	"fmt"
	"github.com/google/uuid"
	"time"
)

type TransferHistory struct {
	AccountFrom    uuid.UUID `json:"accountFrom"`
	AccountTo      uuid.UUID `json:"accountTo"`
	Amount         float64   `json:"amount"`
	TimeOfCreation time.Time
}

type Transaction struct {
	AccountFrom uuid.UUID
	AccountTo   uuid.UUID
	ItemPrice   float64
	ItemUUID    uuid.UUID
	IsFrozen    bool
	IsAccepted  bool
}

type TransactionHistory struct {
	TransactionId  uuid.UUID
	TimeOfCreation time.Time
	// Operation type have two options: ('CT', 'CREATED'), ('CD', 'COMPLETED')
	OperationType string
}

func (t *TransferHistory) String() string {
	return fmt.Sprintf(
		`Account from: %s
				Account to: %s
				Time of creation: %s`,
		t.AccountFrom.String(),
		t.AccountTo.String(),
		t.TimeOfCreation.String(),
	)
}

func (t *Transaction) String() string {
	return fmt.Sprintf(
		`Account from: %s
				Account to: %s
				Item_UUID: %s`,
		t.AccountFrom.String(),
		t.AccountTo.String(),
		t.ItemUUID.String(),
	)
}

func (t *TransactionHistory) String() string {
	return fmt.Sprintf(
		`Transaction_ID: %s
				Time of creation: %s`,
		t.TransactionId.String(),
		t.TimeOfCreation.String(),
	)
}
