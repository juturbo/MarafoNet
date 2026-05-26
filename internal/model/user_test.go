package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUser(t *testing.T) {
	name := "testuser"
	password := "testpassword123"

	user := NewUser(name, password)

	assert.Equal(t, name, user.Name)
	assert.Equal(t, password, user.Password)
}

func TestNewUser_EmptyName(t *testing.T) {
	name := ""
	password := "password"
	user := NewUser(name, password)
	assert.Equal(t, name, user.Name)
}

func TestNewUser_EmptyPassword(t *testing.T) {
	name := "username"
	password := ""
	user := NewUser(name, password)
	assert.Equal(t, password, user.Password)
}

func TestGeneratePasswordHash(t *testing.T) {
	name := "testuser"
	password := "mypassword"
	user := NewUser(name, password)

	hash, err := user.GeneratePasswordHash()
	require.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, password, hash)
}

func TestGeneratePasswordHash_Unique(t *testing.T) {
	name := "testuser"
	password := "samepassword"
	user := NewUser(name, password)

	hash1, err := user.GeneratePasswordHash()
	require.NoError(t, err)

	hash2, err := user.GeneratePasswordHash()
	require.NoError(t, err)

	assert.NotEqual(t, hash1, hash2)
}

func TestGeneratePasswordHash_EmptyPassword(t *testing.T) {
	name := "testuser"
	password := ""
	user := NewUser(name, password)

	hash, err := user.GeneratePasswordHash()
	require.NoError(t, err)
	assert.NotEmpty(t, hash)
}

func TestCheckPassword_Correct(t *testing.T) {
	name := "testuser"
	password := "correctpassword123"
	user := NewUser(name, password)

	hash, err := user.GeneratePasswordHash()
	require.NoError(t, err)

	result := user.CheckPassword(hash)
	assert.True(t, result)
}

func TestCheckPassword_Wrong(t *testing.T) {
	name := "testuser"
	password := "correctpassword"
	user := NewUser(name, password)

	correctHash, err := user.GeneratePasswordHash()
	require.NoError(t, err)

	wrongPassword := "wrongpassword"
	wrongUser := NewUser(name, wrongPassword)
	result := wrongUser.CheckPassword(correctHash)
	assert.False(t, result)
}

func TestCheckPassword_EmptyHash(t *testing.T) {
	name := "testuser"
	password := "password"
	user := NewUser(name, password)

	tryEmptyHash := ""
	result := user.CheckPassword(tryEmptyHash)
	assert.False(t, result)
}

func TestCheckPassword_InvalidHash(t *testing.T) {
	name := "testuser"
	password := "password"
	user := NewUser(name, password)
	result := user.CheckPassword("invalid_hash_format")
	assert.False(t, result)
}

func TestPasswordFlow_Roundtrip(t *testing.T) {
	name := "alice"
	password := "my_secure_password_123"
	user := NewUser(name, password)

	hash, err := user.GeneratePasswordHash()
	require.NoError(t, err)

	sameUser := NewUser(name, password)
	assert.True(t, sameUser.CheckPassword(hash))

	differentPassword := "different_password"
	differentUser := NewUser(name, differentPassword)
	assert.False(t, differentUser.CheckPassword(hash))
}

func TestMultipleUsers_IndependentPasswords(t *testing.T) {
	name1 := "alice"
	password1 := "alice_password"
	user1 := NewUser(name1, password1)

	name2 := "bob"
	password2 := "bob_password"
	user2 := NewUser(name2, password2)

	hash1, _ := user1.GeneratePasswordHash()
	hash2, _ := user2.GeneratePasswordHash()

	assert.True(t, user1.CheckPassword(hash1))
	assert.False(t, user1.CheckPassword(hash2))

	assert.True(t, user2.CheckPassword(hash2))
	assert.False(t, user2.CheckPassword(hash1))
}

func TestSpecialCharactersInPassword(t *testing.T) {
	name := "testuser"
	specialPassword := "P@$$w0rd!#%&*()[]{}|<>?,.;:'\""
	user := NewUser(name, specialPassword)

	hash, err := user.GeneratePasswordHash()
	require.NoError(t, err)

	result := user.CheckPassword(hash)
	assert.True(t, result)
}

func TestUserNameWithSpecialChars(t *testing.T) {
	name := "user@example.com"
	password := "password123"
	user := NewUser(name, password)
	hash, err := user.GeneratePasswordHash()
	require.NoError(t, err)

	assert.True(t, user.CheckPassword(hash))
}

func TestPasswordNotStoredPlaintext(t *testing.T) {
	name := "testuser"
	password := "secretpassword"
	user := NewUser(name, password)

	hash, _ := user.GeneratePasswordHash()

	assert.NotContains(t, hash, password)
}
