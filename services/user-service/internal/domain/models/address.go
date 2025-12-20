package models

import "time"

type AddressType string

const (
	AddressTypeHome  AddressType = "home"
	AddressTypeWork  AddressType = "work"
	AddressTypeOther AddressType = "other"
)

type Address struct {
	ID           string
	UserID       string
	AddressType  AddressType
	FullName     string
	Phone        string
	AddressLine1 string
	AddressLine2 string
	Ward         string
	City         string
	Country      string
	IsDefault    bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
