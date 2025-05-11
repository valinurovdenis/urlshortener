package userstorage

import (
	"context"
	"fmt"
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
	userID := int64(1)
	mock.ExpectQuery("INSERT").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(userID))
	uuid, err := storage.GenerateUUID(context.Background())
	require.Equal(t, uuid, userID)
	require.NoError(t, err)
}

func TestDatabaseUserStorage_init(t *testing.T) {
	// Error in commit
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	// Error in transaction
	storage := &DatabaseUserStorage{DB: db}
	mock.ExpectBegin().WillReturnError(fmt.Errorf("some strange error"))
	err = storage.init()
	require.Error(t, err)

	// ok
	mock.ExpectBegin()
	mock.ExpectExec("CREATE")
	mock.ExpectCommit()
	err = storage.init()
	require.NoError(t, err)
}
