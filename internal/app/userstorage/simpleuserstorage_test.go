package userstorage_test

import (
	"context"
	"testing"

	"github.com/valinurovdenis/urlshortener/internal/app/userstorage"
)

func TestSimpleUserStorage_GenerateUUID(t *testing.T) {
	var userStorage = userstorage.NewSimpleUserStorage()
	ch := make(chan int64)
	var getNewUUID = func() {
		uuid, err := userStorage.GenerateUUID(context.Background())
		if err != nil {
			panic(err)
		}
		ch <- uuid
	}

	const numberOfThreads = 10
	for i := 0; i < numberOfThreads; i++ {
		go getNewUUID()
	}

	var uuids = make(map[int64]struct{})
	for i := 0; i < numberOfThreads; i++ {
		uuid := <-ch
		_, has := uuids[uuid]
		if has {
			t.Errorf("uuid %d is duplicate", uuid)
		}
	}
}
