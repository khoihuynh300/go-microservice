package convert

func PtrIfValid[T any](v T, valid bool) *T {
	if !valid {
		return nil
	}
	return &v
}
