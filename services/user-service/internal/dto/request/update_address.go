package request

type UpdateAddressRequest struct {
	AddressType  *string
	FullName     *string
	Phone        *string
	AddressLine1 *string
	AddressLine2 *string
	Ward         *string
	City         *string
	Country      *string
	IsDefault    *bool
}
