package main

import (
	"bytes"
	"context"
	"log"
	"testing"

	"github.com/im-wmkong/gorm-query/example/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestMain_LogsFoundActiveUsers(t *testing.T) {
	var buf bytes.Buffer
	originalWriter := log.Writer()
	originalFlags := log.Flags()
	originalPrefix := log.Prefix()

	log.SetOutput(&buf)
	log.SetFlags(0)
	log.SetPrefix("")
	defer func() {
		log.SetOutput(originalWriter)
		log.SetFlags(originalFlags)
		log.SetPrefix(originalPrefix)
	}()

	main()

	assert.Contains(t, buf.String(), "Found active users: 1")
	assert.NotContains(t, buf.String(), "failed")
}

func TestMain_LogsFailuresWhenSharedMemoryDBIsLocked(t *testing.T) {
	lockerDB, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, lockerDB.AutoMigrate(&model.User{}))

	tx := lockerDB.Begin()
	require.NoError(t, tx.Error)
	require.NoError(t, tx.WithContext(context.Background()).Create(&model.User{
		UserName: "Locker",
		Age:      20,
		Email:    "locker@example.com",
	}).Error)
	defer func() {
		_ = tx.Rollback().Error
	}()

	var buf bytes.Buffer
	originalWriter := log.Writer()
	originalFlags := log.Flags()
	originalPrefix := log.Prefix()

	log.SetOutput(&buf)
	log.SetFlags(0)
	log.SetPrefix("")
	defer func() {
		log.SetOutput(originalWriter)
		log.SetFlags(originalFlags)
		log.SetPrefix(originalPrefix)
	}()

	main()

	assert.Contains(t, buf.String(), "CreateUser failed:")
	assert.Contains(t, buf.String(), "GetActiveUsers failed:")
}
