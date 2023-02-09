package db

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/longht077/simplebank/utils"
	"github.com/stretchr/testify/require"
)

func CreateRandomTestEntry(t *testing.T) Entry {
	account := createRandomAccount(t)

	arg := CreateEntryParams{
		AccountID: account.ID,
		Ammount:   utils.RandomInt(1, 10000),
	}
	entry, err := testQueries.CreateEntry(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, entry)

	require.Equal(t, arg.AccountID, entry.AccountID)
	require.Equal(t, arg.Ammount, entry.Ammount)

	require.NotZero(t, entry.AccountID)
	require.NotZero(t, entry.Ammount)
	return entry
}

func TestCreateEntry(t *testing.T) {
	CreateRandomTestEntry(t)
}

func TestGetEntry(t *testing.T) {
	entryRandom := CreateRandomTestEntry(t)

	entry, err := testQueries.GetEntry(context.Background(), entryRandom.ID)

	require.NoError(t, err)
	require.NotEmpty(t, entry)

	require.Equal(t, entryRandom.ID, entry.ID)
	require.Equal(t, entryRandom.AccountID, entry.AccountID)
	require.Equal(t, entryRandom.Ammount, entry.Ammount)
	require.WithinDuration(t, entryRandom.CreatedAt, entry.CreatedAt, time.Second)
}

func TestUpdateEntry(t *testing.T) {
	entryRandom := CreateRandomTestEntry(t)

	arg := UpdateEntryParams{
		ID:      entryRandom.ID,
		Ammount: utils.RandomBalance(),
	}

	updatedEntry, err := testQueries.UpdateEntry(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, updatedEntry)

	require.Equal(t, entryRandom.ID, updatedEntry.ID)
	require.Equal(t, arg.Ammount, updatedEntry.Ammount)
	require.Equal(t, entryRandom.AccountID, updatedEntry.AccountID)
	require.Equal(t, entryRandom.CreatedAt, updatedEntry.CreatedAt, time.Second)
}

func TestDeleteEntry(t *testing.T) {
	entryRandom := CreateRandomTestEntry(t)

	err := testQueries.DeleteEntry(context.Background(), entryRandom.ID)

	require.NoError(t, err)

	entry, err := testQueries.GetEntry(context.Background(), entryRandom.ID)

	require.Error(t, err)
	require.EqualError(t, err, sql.ErrNoRows.Error())
	require.Empty(t, entry)
}

func TestListEntries(t *testing.T) {
	for i := 0; i < 10; i++ {
		CreateRandomTestEntry(t)
	}

	arg := ListEntriesParams{
		Limit:  5,
		Offset: 5,
	}

	entries, err := testQueries.ListEntries(context.Background(), arg)

	require.NoError(t, err)
	require.Len(t, entries, 5)

	for _, entry := range entries {
		require.NotEmpty(t, entry)
	}
}
