package urlstorage_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valinurovdenis/urlshortener/internal/app/urlstorage"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestDatabaseStorage_GetLongURL(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	storage := urlstorage.NewDatabaseStorage(db)
	tests := []struct {
		name     string
		s        *urlstorage.DatabaseStorage
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
				require.EqualError(t, err, urlstorage.ErrDeletedURL.Error())
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

	storage := urlstorage.NewDatabaseStorage(db)
	tests := []struct {
		name    string
		s       *urlstorage.DatabaseStorage
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

	storage := urlstorage.NewDatabaseStorage(db)
	tests := []struct {
		name          string
		longURL       string
		shortURL      string
		expectedError error
	}{
		{name: "store_b", longURL: "url_b", shortURL: "b", expectedError: nil},
		{name: "store_empty", longURL: "", shortURL: "", expectedError: urlstorage.ErrEmptyLongURL},
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

	storage := urlstorage.NewDatabaseStorage(db)
	tests := []struct {
		name          string
		urlsToStore   []urlstorage.URLPair
		expectedError error
	}{
		{name: "store_many",
			urlsToStore: []urlstorage.URLPair{
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

	storage := urlstorage.NewDatabaseStorage(db)
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
				require.Equal(t, []urlstorage.URLPair{{Long: "url_a", Short: "a"}}, rows)
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

	storage := urlstorage.NewDatabaseStorage(db)
	mock.ExpectExec("UPDATE shortener SET deleted=true").WillReturnResult(sqlmock.NewResult(1, 1))
	storage.DeleteUserURLs(context.Background(), urlstorage.URLsForDelete{UserID: "user_1", ShortURLs: []string{"a", "b"}})
}
