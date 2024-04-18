package db_tests

import (
	"context"
	"testing"

	db "github.com/koliader/posts-auth.git/internal/db/sqlc"
	"github.com/koliader/posts-auth.git/internal/util"
	"github.com/stretchr/testify/require"
)

func createRandomUser(t *testing.T) db.User {
	password := util.RandomString(5)
	hashedPassword, err := util.HashPassword(password)
	require.NoError(t, err)

	arg := db.CreateUserParams{
		Email:    util.RandomEmail(),
		Username: util.RandomString(5),
		Password: hashedPassword,
	}
	user, err := testStore.CreateUser(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, user)

	require.Equal(t, arg.Email, user.Email)
	require.Equal(t, arg.Password, user.Password)
	require.Equal(t, arg.Username, user.Username)
	return user
}

func TestCreateUser(t *testing.T) {
	createRandomUser(t)
}

func TestGetUserByEmail(t *testing.T) {
	user1 := createRandomUser(t)
	user2, err := testStore.GetUserByEmail(context.Background(), user1.Email)
	require.NoError(t, err)
	require.NotEmpty(t, user2)

	require.Equal(t, user1.Email, user2.Email)
	require.Equal(t, user1.Password, user2.Password)
	require.Equal(t, user1.Username, user2.Username)
}

func TestListUsers(t *testing.T) {
	for i := 0; i < 5; i++ {
		createRandomUser(t)
	}
	users, err := testStore.ListUsers(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, users)
}

func TestUpdateUserEmail(t *testing.T) {
	user1 := createRandomUser(t)
	arg := db.UpdateUserEmailParams{
		Email:   user1.Email,
		Email_2: util.RandomEmail(),
	}
	user2, err := testStore.UpdateUserEmail(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, user2)

	require.Equal(t, arg.Email_2, user2.Email)
}
