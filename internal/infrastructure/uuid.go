package infra

import (
	gonanoid "github.com/matoous/go-nanoid"
)

type UUIDGenerator interface {
	Generate() (string, error)
}

type NanoIDGenerator struct {
	Length int
}

func NewNanoIDGenerator(length int) *NanoIDGenerator {
	if length < 1 {
		panic("length must be larger than 1")
	}
	return &NanoIDGenerator{Length: length}
}

func (ns *NanoIDGenerator) Generate() (string, error) {
	uuid, err := gonanoid.Nanoid(ns.Length)
	if err != nil {
		return "", err
	}
	return uuid, err
}
