package convert

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

var (
	DateFormat = "02-01-2006"
)

func TimestampToTimePtr(ts *timestamppb.Timestamp) *time.Time {
	if ts == nil {
		return nil
	}
	t := ts.AsTime()
	return &t
}

func TimePtrToTimestamp(t *time.Time) *timestamppb.Timestamp {
	if t == nil {
		return nil
	}
	return timestamppb.New(*t)
}

func TimePtrToDateStringWrapper(t *time.Time) *wrapperspb.StringValue {
	if t == nil {
		return nil
	}
	return wrapperspb.String(t.Format(DateFormat))
}

func StringPtrToTimePtr(dateStr *string) (*time.Time, error) {
	if dateStr == nil {
		return nil, nil
	}
	t, err := time.Parse(DateFormat, *dateStr)
	if err != nil {
		return nil, fmt.Errorf("invalid date format, expected DD-MM-YYYY: %w", err)
	}
	return &t, nil
}

func GenericStringPtrToWrapper[T ~string](val *T) *wrapperspb.StringValue {
	if val == nil {
		return nil
	}
	return wrapperspb.String(string(*val))
}

func StringWrapperToPtr(s *wrapperspb.StringValue) *string {
	if s == nil {
		return nil
	}

	v := s.Value
	return &v
}

func PtrToStringWrapper(s *string) *wrapperspb.StringValue {
	if s == nil {
		return nil
	}
	return wrapperspb.String(*s)
}

func PtrUUIDToStringWrapper(id *uuid.UUID) *wrapperspb.StringValue {
	if id == nil {
		return nil
	}

	return wrapperspb.String(id.String())
}

func DoubleWrapperToPtr(d *wrapperspb.DoubleValue) *float64 {
	if d == nil {
		return nil
	}

	v := d.Value
	return &v
}
