package settings

import (
	"context"
	"errors"
	
	"github.com/monarch-dev/monarch/database"
	"github.com/monarch-dev/monarch/internal/crypto"
	"github.com/jackc/pgx/v5/pgtype"
)

type Service struct {
	q      *database.Queries
	encKey []byte
}

func NewService(db database.DBTX, encKey []byte) *Service {
	return &Service{
		q:      database.New(db),
		encKey: encKey,
	}
}

func (s *Service) Set(ctx context.Context, key string, value string, encrypt bool) error {
	var data []byte
	if encrypt {
		if len(s.encKey) != 32 {
			return errors.New("invalid encryption key length")
		}
		var err error
		data, err = crypto.Encrypt([]byte(value), s.encKey)
		if err != nil {
			return err
		}
	} else {
		data = []byte(value)
	}
	
	return s.q.UpsertSetting(ctx, database.UpsertSettingParams{
		Key:         key,
		Value:       data,
		IsEncrypted: pgtype.Bool{Bool: encrypt, Valid: true},
	})
}

func (s *Service) Get(ctx context.Context, key string) (string, error) {
	row, err := s.q.GetSetting(ctx, key)
	if err != nil {
		return "", err
	}
	
	if row.IsEncrypted.Bool {
		decrypted, err := crypto.Decrypt(row.Value, s.encKey)
		if err != nil {
			return "", err
		}
		return string(decrypted), nil
	}
	
	return string(row.Value), nil
}
