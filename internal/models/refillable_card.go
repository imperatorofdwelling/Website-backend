package models

import (
	"errors"
	"strings"
)

// (Yan) TODO payment card struct

const (
	CardSize int = 16
)

type RefillableCard struct {
	Owner    *User  `json:"owner"`
	Synonym  string `json:"synonymic"`
	CardMask string `json:"cardMask"`
}

func NewRefillableCard(owner *User, synonym, cardMask string) *RefillableCard {
	return &RefillableCard{
		Owner:    owner,
		Synonym:  synonym,
		CardMask: cardMask,
	}
}

func (card *RefillableCard) IsFullData() bool {
	if card == nil {
		return false
	}
	if card.Owner == nil || len(card.Synonym) == 0 || len(card.CardMask) == 0 {
		return false
	}
	return true
}

func GenerateCardMask(firstSix string, lastFour string) (string, error) {
	if len(firstSix) != 6 || len(lastFour) != 4 {
		return "", errors.New("first and Last length must be 6 or 4")
	}
	missedNumbers := strings.Repeat("*", CardSize-6-4)
	return firstSix + missedNumbers + lastFour, nil
}
