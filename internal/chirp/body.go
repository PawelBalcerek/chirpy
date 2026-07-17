package chirp

import (
	"errors"
	"strings"
)

var (
	ErrBodyTooLong = errors.New("Chirp is too long")
	ErrBodyEmpty   = errors.New("Chirp body cannot be empty")
)

var profaneWords = map[string]struct{}{
	"kerfuffle": {},
	"sharbert":  {},
	"fornax":    {},
}

type Body struct {
	value string
}

func NewBody(raw string) (Body, error) {
	if strings.TrimSpace(raw) == "" {
		return Body{}, ErrBodyEmpty
	}
	if len([]rune(raw)) > 140 {
		return Body{}, ErrBodyTooLong
	}

	words := strings.Split(raw, " ")
	for i, word := range words {
		if _, ok := profaneWords[strings.ToLower(word)]; ok {
			words[i] = "****"
		}
	}
	cleaned := strings.Join(words, " ")

	return Body{value: cleaned}, nil
}

func (b Body) String() string {
	return b.value
}
