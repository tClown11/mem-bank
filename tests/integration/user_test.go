package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"mem_bank/internal/domain"
	"mem_bank/internal/repository"
	"mem_bank/internal/usecase"
	"mem_bank/tests/testutil"
)

type UserTestSuite struct {
	suite.Suite
	db          *testutil.TestDB
	userRepo    domain.UserRepository
	memoryRepo  domain.MemoryRepository
	userUsecase usecase.UserUsecase
}

func (suite *UserTestSuite) SetupSuite() {
	db, err := testutil.SetupTestDB()
	if err != nil {
		suite.T().Fatalf("Failed to setup test database: %v", err)
	}
	suite.db = db

	suite.userRepo = repository.NewUserRepository(db.Pool)
	suite.memoryRepo = repository.NewMemoryRepository(db.Pool)
	suite.userUsecase = usecase.NewUserUsecase(suite.userRepo, suite.memoryRepo)
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
	req := &domain.UserCreateRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Profile: domain.UserProfile{
			FirstName: "Test",
			LastName:  "User",
		},
		Settings: domain.UserSettings{
			Language: "en",
			Timezone: "UTC",
		},
	}

	user, err := suite.userUsecase.CreateUser(req)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), user)
	assert.Equal(suite.T(), req.Username, user.Username)
	assert.Equal(suite.T(), req.Email, user.Email)
	assert.True(suite.T(), user.IsActive)
}

func (suite *UserTestSuite) TestCreateUserDuplicate() {
	req := &domain.UserCreateRequest{
		Username: "testuser",
		Email:    "test@example.com",
	}

	_, err := suite.userUsecase.CreateUser(req)
	assert.NoError(suite.T(), err)

	_, err = suite.userUsecase.CreateUser(req)
	assert.Equal(suite.T(), domain.ErrUserAlreadyExists, err)
}

func (suite *UserTestSuite) TestGetUserByID() {
	req := &domain.UserCreateRequest{
		Username: "testuser",
		Email:    "test@example.com",
	}

	createdUser, err := suite.userUsecase.CreateUser(req)
	assert.NoError(suite.T(), err)

	user, err := suite.userUsecase.GetUserByID(createdUser.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), createdUser.ID, user.ID)
	assert.Equal(suite.T(), createdUser.Username, user.Username)
}

func (suite *UserTestSuite) TestUpdateUser() {
	req := &domain.UserCreateRequest{
		Username: "testuser",
		Email:    "test@example.com",
	}

	createdUser, err := suite.userUsecase.CreateUser(req)
	assert.NoError(suite.T(), err)

	newUsername := "updateduser"
	updateReq := &domain.UserUpdateRequest{
		Username: &newUsername,
	}

	updatedUser, err := suite.userUsecase.UpdateUser(createdUser.ID, updateReq)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), newUsername, updatedUser.Username)
}

func TestUserSuite(t *testing.T) {
	suite.Run(t, new(UserTestSuite))
}