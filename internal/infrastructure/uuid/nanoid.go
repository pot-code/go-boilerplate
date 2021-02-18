package uuid

import gonanoid "github.com/matoous/go-nanoid"

// Generator UUID generator interface
type Generator interface {
	Generate() (string, error)
}

// NanoIDGenerator UUID implementation using NanoID
type NanoIDGenerator struct {
	Length int
}

var _ Generator = &NanoIDGenerator{}

// NewNanoIDGenerator create a new `NanoIDGenerator` instance
func NewNanoIDGenerator(length int) *NanoIDGenerator {
	if length < 1 {
		panic("length must be larger than 1")
	}
	return &NanoIDGenerator{Length: length}
}

// Generate generate UUID
func (ns *NanoIDGenerator) Generate() (string, error) {
	uuid, err := gonanoid.Nanoid(ns.Length)
	if err != nil {
		return "", err
	}
	return uuid, err
}
