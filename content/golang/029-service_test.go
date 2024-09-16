package services_test

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/moroz/webauthn-academy-go/db/queries"
	"github.com/stretchr/testify/suite"
)

type ServiceTestSuite struct {
	suite.Suite
	db queries.DBTX
}

func TestServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}

func (s *ServiceTestSuite) SetupTest() {
	connString := os.Getenv("TEST_DATABASE_URL")
	db, err := pgx.Connect(context.Background(), connString)
	s.NoError(err)
	s.db = db

	_, err = s.db.Exec(context.Background(), "truncate users")
	s.NoError(err)
}

