package urlstorage

import (
	"context"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDatabaseStorage_GetLongURL(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	storage := NewDatabaseStorage(db)
	tests := []struct {
		name     string
		s        *DatabaseStorage
		shortURL string
		want     string
		wantErr  bool
		deleted  bool
	}{
		{name: "get_a", s: storage, shortURL: "a", want: "url_a", wantErr: false, deleted: false},
		{name: "get_b", s: storage, shortURL: "b", want: "", wantErr: true, deleted: true},
		{name: "get_empty", s: storage, shortURL: "", want: "", wantErr: true, deleted: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr && !tt.deleted {
				mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{}))
			} else {
				mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"long_url", "deleted"}).AddRow(tt.want, tt.deleted))
			}
			got, err := tt.s.GetLongURLWithContext(context.Background(), tt.shortURL)
			if tt.deleted {
				require.EqualError(t, err, ErrDeletedURL.Error())
			} else if !tt.wantErr {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDatabaseStorage_GetShortURL(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	storage := NewDatabaseStorage(db)
	tests := []struct {
		name    string
		s       *DatabaseStorage
		longURL string
		want    string
		wantErr bool
	}{
		{name: "get_a", s: storage, longURL: "url_a", want: "a", wantErr: false},
		{name: "get_b", s: storage, longURL: "url_b", want: "", wantErr: true},
		{name: "get_empty", s: storage, longURL: "", want: "", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{}))
			} else {
				mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"short_url"}).AddRow(tt.want))
			}
			got, err := tt.s.GetShortURLWithContext(context.Background(), tt.longURL)
			if !tt.wantErr {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDatabaseStorage_Store(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	storage := NewDatabaseStorage(db)
	tests := []struct {
		name          string
		longURL       string
		shortURL      string
		expectedError error
	}{
		{name: "store_b", longURL: "url_b", shortURL: "b", expectedError: nil},
		{name: "store_empty", longURL: "", shortURL: "", expectedError: ErrEmptyLongURL},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectedError == nil {
				mock.ExpectExec("INSERT").WillReturnResult(sqlmock.NewResult(1, 1))
			}
			err := storage.StoreWithContext(context.Background(), tt.longURL, tt.shortURL, "")
			require.Equal(t, tt.expectedError, err)

		})
	}
}

func TestDatabaseStorage_StoreMany(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	storage := NewDatabaseStorage(db)
	tests := []struct {
		name          string
		urlsToStore   []URLPair
		expectedError error
	}{
		{name: "store_many",
			urlsToStore: []URLPair{
				{Long: "url_a", Short: "a"}},
			expectedError: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock.ExpectBegin()
			mock.ExpectPrepare("INSERT INTO shortener").ExpectExec().WillReturnResult(sqlmock.NewResult(1, 1))
			mock.ExpectCommit()
			errs, err := storage.StoreManyWithContext(context.Background(), tt.urlsToStore, "")
			require.Equal(t, tt.expectedError, err)
			assert.Equal(t, errs, []error{nil})
		})
	}
}

func TestDatabaseStorage_GetUserURLs(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	storage := NewDatabaseStorage(db)
	tests := []struct {
		name      string
		user      string
		wantError bool
	}{
		{name: "empty", user: "user_1", wantError: true},
		{name: "not_empty", user: "user_2", wantError: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantError {
				mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"short_url", "long_url"}))
			} else {
				mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"short_url", "long_url"}).AddRow("a", "url_a"))
			}
			rows, err := storage.GetUserURLs(context.Background(), tt.user)
			if !tt.wantError {
				require.NoError(t, err)
				require.Equal(t, []URLPair{{Long: "url_a", Short: "a"}}, rows)
			}

		})
	}

	storage.GetUserURLs(context.Background(), "user_1")
}

func TestDatabaseStorage_DeleteUserURLs(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	storage := NewDatabaseStorage(db)
	urlsForDelete := URLsForDelete{UserID: "user_1", ShortURLs: []string{"a", "b"}}

	// OK
	mock.ExpectBegin()
	mock.ExpectPrepare("UPDATE")
	mock.ExpectExec("UPDATE shortener SET deleted=true").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	err = storage.DeleteUserURLs(context.Background(), urlsForDelete)
	require.NoError(t, err)

	// Error in commit
	mock.ExpectBegin().WillReturnError(fmt.Errorf("some strange error"))
	err = storage.DeleteUserURLs(context.Background(), urlsForDelete)
	require.Error(t, err)

	// Error in prepare
	mock.ExpectPrepare("update").WillReturnError(fmt.Errorf("some strange error"))
	err = storage.DeleteUserURLs(context.Background(), urlsForDelete)
	require.Error(t, err)
}

func TestDatabaseStorage_GetStats(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	storage := NewDatabaseStorage(db)
	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"userCount", "urlCount"}).AddRow(5, 3))
	res, err := storage.GetStats(context.Background())
	require.NoError(t, err)
	require.Equal(t, StorageStats{URLCount: 3, UserCount: 5}, res)
}

func TestDatabaseStorage_init(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec("CREATE TABLE shortener").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("CREATE INDEX user_id_index").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("CREATE UNIQUE INDEX long_url_index").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	storage := NewDatabaseStorage(db)
	storage.init()
}

func TestDatabaseStorage_Clear(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	storage := NewDatabaseStorage(db)

	// ok
	mock.ExpectBegin()
	mock.ExpectExec("delete").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	err = storage.Clear()
	require.NoError(t, err)

	// Error in begin
	mock.ExpectBegin().WillReturnError(fmt.Errorf("some strange error"))
	err = storage.Clear()
	require.Error(t, err)
}

func TestDatabaseStorage_Ping(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectCommit()

	storage := NewDatabaseStorage(db)
	storage.Ping()
}
