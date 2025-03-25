package userstorage

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestDatabaseUserStorage_GenerateUUID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	storage := NewDatabaseUserStorage(db)
	userId := int64(1)
	mock.ExpectQuery("INSERT").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(userId))
	uuid, err := storage.GenerateUUID(context.Background())
	require.Equal(t, uuid, userId)
	require.NoError(t, err)
}
