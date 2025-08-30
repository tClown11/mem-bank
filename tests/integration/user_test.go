package integration

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	userDao "mem_bank/internal/dao/user"
	"mem_bank/internal/domain/user"
	userService "mem_bank/internal/service/user"
	"mem_bank/tests/testutil"
)

type UserTestSuite struct {
	suite.Suite
	db          *testutil.TestDB
	userService user.Service
	ctx         context.Context
}

func (suite *UserTestSuite) SetupSuite() {
	db, err := testutil.SetupTestDB()
	if err != nil {
		suite.T().Fatalf("Failed to setup test database: %v", err)
	}
	suite.db = db

	// Initialize service with new architecture
	userRepo := userDao.NewPostgresRepository(db.GormDB)
	suite.userService = userService.NewService(userRepo)
	suite.ctx = context.Background()
}

func (suite *UserTestSuite) TearDownSuite() {
	if suite.db != nil {
		suite.db.Close()
	}
}

func (suite *UserTestSuite) SetupTest() {
	suite.db.CleanupTables()
}

func (suite *UserTestSuite) TestCreateUser() {
	req := user.CreateRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Profile: user.Profile{
			FirstName: "Test",
			LastName:  "User",
		},
		Settings: user.Settings{
			Language: "en",
			Timezone: "UTC",
		},
	}

	u, err := suite.userService.CreateUser(suite.ctx, req)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), u)
	assert.Equal(suite.T(), req.Username, u.Username)
	assert.Equal(suite.T(), req.Email, u.Email)
	assert.True(suite.T(), u.IsActive)
}

func (suite *UserTestSuite) TestCreateUserDuplicate() {
	req := user.CreateRequest{
		Username: "testuser",
		Email:    "test@example.com",
	}

	_, err := suite.userService.CreateUser(suite.ctx, req)
	assert.NoError(suite.T(), err)

	_, err = suite.userService.CreateUser(suite.ctx, req)
	assert.Equal(suite.T(), user.ErrUsernameTaken, err)
}

func (suite *UserTestSuite) TestGetUserByID() {
	req := user.CreateRequest{
		Username: "testuser",
		Email:    "test@example.com",
	}

	createdUser, err := suite.userService.CreateUser(suite.ctx, req)
	assert.NoError(suite.T(), err)

	u, err := suite.userService.GetUser(suite.ctx, createdUser.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), createdUser.ID, u.ID)
	assert.Equal(suite.T(), createdUser.Username, u.Username)
}

func (suite *UserTestSuite) TestUpdateUser() {
	req := user.CreateRequest{
		Username: "testuser",
		Email:    "test@example.com",
	}

	createdUser, err := suite.userService.CreateUser(suite.ctx, req)
	assert.NoError(suite.T(), err)

	newUsername := "updateduser"
	updateReq := user.UpdateRequest{
		Username: &newUsername,
	}

	updatedUser, err := suite.userService.UpdateUser(suite.ctx, createdUser.ID, updateReq)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), newUsername, updatedUser.Username)
}

func TestUserSuite(t *testing.T) {
	suite.Run(t, new(UserTestSuite))
}
