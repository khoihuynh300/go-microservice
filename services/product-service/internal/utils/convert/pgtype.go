package convert

import (
	"strconv"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func PtrToUUID(p *uuid.UUID) pgtype.UUID {
	if p == nil {
		return pgtype.UUID{}
	}
	return pgtype.UUID{
		Bytes: *p,
		Valid: true,
	}
}

func PgUUIDToPtr(u pgtype.UUID) *uuid.UUID {
	if !u.Valid {
		return nil
	}

	id, err := uuid.FromBytes(u.Bytes[:])
	if err != nil {
		return nil
	}

	return &id
}

func DoubleToNumeric(f float64) (pgtype.Numeric, error) {
	var numeric pgtype.Numeric
	if err := numeric.Scan(strconv.FormatFloat(f, 'f', 2, 64)); err != nil {
		return numeric, err
	}

	return numeric, nil
}

func NumericToDouble(n pgtype.Numeric) float64 {
	var price float64
	if n.Valid {
		if floatVal, err := n.Float64Value(); err == nil {
			price = floatVal.Float64
		}
	}
	return price
}
