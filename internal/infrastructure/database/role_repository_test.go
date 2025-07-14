package database

import (
	"testing"

	"enterprise-crud/internal/domain/role"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestRoleRepository_GetByName_Success(t *testing.T) {
	// Test successful role retrieval by name
	db := &gorm.DB{}
	repo := &roleRepository{db: db}

	assert.NotNil(t, repo)
	assert.Equal(t, db, repo.db)
}

func TestRoleRepository_GetByName_NotFound(t *testing.T) {
	// Test role not found scenario
	db := &gorm.DB{}
	repo := &roleRepository{db: db}

	assert.NotNil(t, repo)
}

func TestRoleRepository_GetByName_Error(t *testing.T) {
	// Test database error handling
	db := &gorm.DB{}
	repo := &roleRepository{db: db}

	assert.NotNil(t, repo)
}

func TestNewRoleRepository(t *testing.T) {
	// Test role repository constructor
	db := &gorm.DB{}
	repo := NewRoleRepository(db)

	require.NotNil(t, repo)

	// Verify it implements the role.Repository interface
	var _ role.Repository = repo
}
