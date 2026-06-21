package tests

import (
	"books-and-trust/services/user-service/internal/domain"
	"books-and-trust/services/user-service/internal/infra/repo"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSQLRepository_Create_Integration(t *testing.T) {
    ctx := context.Background()
    db, container, err := setupPostgresContainer(ctx)
    assert.NoError(t, err)
    defer func() { _ = container.Terminate(ctx) }()

    err = db.AutoMigrate(&domain.User{})

	repository := repo.NewSQLRepository(db)

	plainPassword := "MamadIntegration123!"
	user := &domain.User{
		Name:     "Integration User",
		Username: "integration_test_user",
		Email:    "integration@test.com",
		Password: domain.Password{
			Text: &plainPassword,
			Hash: []byte("$2a$10$FakeHashForIntegrationTestOnlyJustToVerifyDatabaseField"),
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = repository.Create(ctx, user)

	assert.NoError(t, err)
	
	assert.NotEmpty(t, user.ID) 

	err = db.Exec("DELETE FROM users WHERE username = ?", "integration_test_user").Error
	assert.NoError(t, err)
}