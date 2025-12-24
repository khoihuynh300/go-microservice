package convert

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func PtrToText[T ~string](p *T) pgtype.Text {
	if p == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{
		String: string(*p),
		Valid:  true,
	}
}

func PtrToDate(p *time.Time) pgtype.Date {
	if p == nil {
		return pgtype.Date{}
	}
	return pgtype.Date{
		Time:  *p,
		Valid: true,
	}
}

func PtrToTimestamptz(p *time.Time) pgtype.Timestamptz {
	if p == nil {
		return pgtype.Timestamptz{}
	}
	return pgtype.Timestamptz{
		Time:  *p,
		Valid: true,
	}
}

func PtrIfValid[T any](v T, valid bool) *T {
	if !valid {
		return nil
	}
	return &v
}
