package db

import (
	"context"
	"database/sql"
	"testing"

	"github.com/longht077/simplebank/utils"
	"github.com/stretchr/testify/require"
)

func CreateRandomTranfer(t *testing.T) Tranfer {
	fromAccount := createRandomAccount(t)
	toAccount := createRandomAccount(t)

	arg := CreateTranferParams{
		FromAccountID: fromAccount.ID,
		ToAccountID:   toAccount.ID,
		Ammount:       utils.RandomInt(0, fromAccount.Balance),
	}
	tranfer, err := testQueries.CreateTranfer(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, tranfer)
	require.Equal(t, fromAccount.ID, tranfer.FromAccountID)
	require.Equal(t, toAccount.ID, tranfer.ToAccountID)
	require.Equal(t, arg.Ammount, tranfer.Ammount)
	return tranfer
}

func TestCreateTranfer(t *testing.T) {
	CreateRandomTranfer(t)
}

func TestGetTranfer(t *testing.T) {
	randomTranfer := CreateRandomTranfer(t)

	tranfer, err := testQueries.GetTranfer(context.Background(), randomTranfer.ID)

	require.NoError(t, err)
	require.NotEmpty(t, tranfer)
	require.Equal(t, randomTranfer.ID, tranfer.ID)
	require.Equal(t, randomTranfer.FromAccountID, tranfer.FromAccountID)
	require.Equal(t, randomTranfer.ToAccountID, tranfer.ToAccountID)
	require.Equal(t, randomTranfer.Ammount, tranfer.Ammount)
}

func TestUpdateTranfer(t *testing.T) {

	randomTranfer := CreateRandomTranfer(t)
	fromAccount, fromErr := testQueries.GetAccount(context.Background(), randomTranfer.FromAccountID)
	toAccount, toErr := testQueries.GetAccount(context.Background(), randomTranfer.ToAccountID)

	require.NoError(t, fromErr)
	require.NoError(t, toErr)
	require.NotEmpty(t, fromAccount)
	require.NotEmpty(t, toAccount)

	arg := UpdateTranferParams{
		ID:      randomTranfer.ID,
		Ammount: utils.RandomInt(0, fromAccount.Balance),
	}

	updateTranfer, err := testQueries.UpdateTranfer(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, updateTranfer)
	require.Equal(t, arg.ID, updateTranfer.ID)
	require.Equal(t, fromAccount.ID, updateTranfer.FromAccountID)
	require.Equal(t, toAccount.ID, updateTranfer.ToAccountID)
	require.Equal(t, arg.Ammount, updateTranfer.Ammount)
}

func TestDeleteTranfer(t *testing.T) {
	randomTranfer := CreateRandomTranfer(t)

	err := testQueries.DeleteTranfer(context.Background(), randomTranfer.ID)
	require.NoError(t, err)

	tranfer, err := testQueries.GetTranfer(context.Background(), randomTranfer.ID)

	require.Error(t, err)
	require.EqualError(t, err, sql.ErrNoRows.Error())
	require.Empty(t, tranfer)
}

func TestListTranfers(t *testing.T) {
	for i := 0; i < 10; i++ {
		CreateRandomTranfer(t)
	}

	arg := ListTranfersParams{
		Limit:  5,
		Offset: 5,
	}

	tranfers, err := testQueries.ListTranfers(context.Background(), arg)

	require.NoError(t, err)
	require.Len(t, tranfers, 5)

	for _, tranfer := range tranfers {
		require.NotEmpty(t, tranfer)
	}
}
