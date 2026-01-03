package impl

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	sqlc "github.com/khoihuynh300/go-microservice/user-service/internal/db/generated"
	"github.com/khoihuynh300/go-microservice/user-service/internal/domain/models"
	"github.com/khoihuynh300/go-microservice/user-service/internal/repository"
)

type addressRepository struct {
	baseRepository
}

func NewAddressRepository(db *pgxpool.Pool) repository.AddressRepository {
	return &addressRepository{
		baseRepository: baseRepository{
			db: db,
			q:  sqlc.New(db),
		},
	}
}

func (r *addressRepository) Create(ctx context.Context, address *models.Address) error {
	params := sqlc.CreateAddressParams{
		ID:           uuid.New(),
		UserID:       address.UserID,
		AddressType:  sqlc.AddressTypeEnum(address.AddressType),
		FullName:     address.FullName,
		Phone:        pgtype.Text{String: address.Phone, Valid: true},
		AddressLine1: address.AddressLine1,
		AddressLine2: pgtype.Text{String: address.AddressLine2, Valid: true},
		Ward:         address.Ward,
		City:         address.City,
		Country:      address.Country,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	result, err := r.queries(ctx).CreateAddress(ctx, params)
	if err != nil {
		return err
	}

	address.ID = result.ID
	address.CreatedAt = result.CreatedAt
	address.UpdatedAt = result.UpdatedAt

	return nil
}

func (r *addressRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Address, error) {
	rows, err := r.queries(ctx).ListAddressesByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	var addresses []*models.Address
	for _, row := range rows {
		addresses = append(addresses, mapToAddress(&row))
	}

	return addresses, nil
}

func (r *addressRepository) FindByIDAndUserID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*models.Address, error) {
	params := sqlc.GetAddressByIDAndUserIDParams{
		ID:     id,
		UserID: userID,
	}
	row, err := r.queries(ctx).GetAddressByIDAndUserID(ctx, params)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return mapToAddress(&row), nil
}

func (r *addressRepository) Update(ctx context.Context, address *models.Address) error {
	params := sqlc.UpdateAddressParams{
		ID:           address.ID,
		AddressType:  sqlc.AddressTypeEnum(address.AddressType),
		FullName:     address.FullName,
		Phone:        pgtype.Text{String: address.Phone, Valid: true},
		AddressLine1: address.AddressLine1,
		AddressLine2: pgtype.Text{String: address.AddressLine2, Valid: true},
		Ward:         address.Ward,
		City:         address.City,
		Country:      address.Country,
		UpdatedAt:    time.Now(),
	}

	result, err := r.queries(ctx).UpdateAddress(ctx, params)
	if err != nil {
		return err
	}

	address.UpdatedAt = result.UpdatedAt
	return nil
}

func (r *addressRepository) SetDefaultAddress(ctx context.Context, userID, addressID uuid.UUID) error {
	return r.queries(ctx).SetDefaultAddress(ctx, sqlc.SetDefaultAddressParams{
		ID:     addressID,
		UserID: userID,
	})
}

func (r *addressRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.queries(ctx).DeleteAddress(ctx, id)

}

func mapToAddress(row *sqlc.UserAddress) *models.Address {
	return &models.Address{
		ID:           row.ID,
		UserID:       row.UserID,
		AddressType:  models.AddressType(row.AddressType),
		FullName:     row.FullName,
		Phone:        row.Phone.String,
		AddressLine1: row.AddressLine1,
		AddressLine2: row.AddressLine2.String,
		Ward:         row.Ward,
		City:         row.City,
		Country:      row.Country,
		IsDefault:    row.IsDefault.Bool,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}
}
