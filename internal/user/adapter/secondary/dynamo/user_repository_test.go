package dynamo

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/shotokan/gymtracker-api/internal/user/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// ── Mocks ─────────────────────────────────────────────────────────────────

type mockDynamoClient struct{ mock.Mock }

func (m *mockDynamoClient) GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dynamodb.GetItemOutput), args.Error(1)
}

func (m *mockDynamoClient) PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dynamodb.PutItemOutput), args.Error(1)
}

func (m *mockDynamoClient) UpdateItem(ctx context.Context, params *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dynamodb.UpdateItemOutput), args.Error(1)
}

func (m *mockDynamoClient) Query(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dynamodb.QueryOutput), args.Error(1)
}

// ── Tests ─────────────────────────────────────────────────────────────────

func TestUserRepository_FindByEmail_Success(t *testing.T) {
	t.Parallel()

	client := new(mockDynamoClient)
	repo := NewUserRepository(client, "gym-tracker")

	userRecord := record{
		PK:           "USER#ivan@gym.com",
		SK:           "PROFILE",
		GSI1PK:       "USER_ID#user-123",
		GSI1SK:       "PROFILE",
		ID:           "user-123",
		Email:        "ivan@gym.com",
		PasswordHash: "hashedPassword",
		Name:         "Ivan",
		Active:       true,
		CreatedAt:    "2024-01-15T10:00:00Z",
		UpdatedAt:    "2024-01-15T10:00:00Z",
	}

	item, _ := attributevalue.MarshalMap(userRecord)

	client.On("GetItem", mock.Anything, mock.MatchedBy(func(input *dynamodb.GetItemInput) bool {
		pk, _ := input.Key["PK"].(*types.AttributeValueMemberS)
		sk, _ := input.Key["SK"].(*types.AttributeValueMemberS)
		return *input.TableName == "gym-tracker" &&
			pk.Value == "USER#ivan@gym.com" &&
			sk.Value == "PROFILE"
	})).Return(&dynamodb.GetItemOutput{Item: item}, nil)

	result, err := repo.FindByEmail(context.Background(), "ivan@gym.com")

	require.NoError(t, err)
	assert.Equal(t, "user-123", result.ID)
	assert.Equal(t, "ivan@gym.com", result.Email)
	assert.Equal(t, "Ivan", result.Name)
	assert.True(t, result.Active)
	client.AssertExpectations(t)
}

func TestUserRepository_FindByEmail_NotFound(t *testing.T) {
	t.Parallel()

	client := new(mockDynamoClient)
	repo := NewUserRepository(client, "gym-tracker")

	client.On("GetItem", mock.Anything, mock.Anything).Return(&dynamodb.GetItemOutput{Item: nil}, nil)

	_, err := repo.FindByEmail(context.Background(), "nonexistent@gym.com")

	require.ErrorIs(t, err, domain.ErrNotFound)
	client.AssertExpectations(t)
}

func TestUserRepository_FindByEmail_Error(t *testing.T) {
	t.Parallel()

	client := new(mockDynamoClient)
	repo := NewUserRepository(client, "gym-tracker")

	client.On("GetItem", mock.Anything, mock.Anything).Return(nil, errors.New("dynamo error"))

	_, err := repo.FindByEmail(context.Background(), "ivan@gym.com")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "dynamo get item")
	client.AssertExpectations(t)
}

func TestUserRepository_FindByID_Success(t *testing.T) {
	t.Parallel()

	client := new(mockDynamoClient)
	repo := NewUserRepository(client, "gym-tracker")

	userRecord := record{
		PK:           "USER#ivan@gym.com",
		SK:           "PROFILE",
		GSI1PK:       "USER_ID#user-123",
		GSI1SK:       "PROFILE",
		ID:           "user-123",
		Email:        "ivan@gym.com",
		PasswordHash: "hashedPassword",
		Name:         "Ivan",
		Active:       true,
		CreatedAt:    "2024-01-15T10:00:00Z",
		UpdatedAt:    "2024-01-15T10:00:00Z",
	}

	item, _ := attributevalue.MarshalMap(userRecord)

	client.On("Query", mock.Anything, mock.MatchedBy(func(input *dynamodb.QueryInput) bool {
		return *input.TableName == "gym-tracker" &&
			*input.IndexName == "GSI1" &&
			*input.KeyConditionExpression == "GSI1PK = :pk AND GSI1SK = :sk" &&
			*input.Limit == 1
	})).Return(&dynamodb.QueryOutput{Items: []map[string]types.AttributeValue{item}}, nil)

	result, err := repo.FindByID(context.Background(), "user-123")

	require.NoError(t, err)
	assert.Equal(t, "user-123", result.ID)
	assert.Equal(t, "ivan@gym.com", result.Email)
	assert.Equal(t, "Ivan", result.Name)
	client.AssertExpectations(t)
}

func TestUserRepository_FindByID_NotFound(t *testing.T) {
	t.Parallel()

	client := new(mockDynamoClient)
	repo := NewUserRepository(client, "gym-tracker")

	client.On("Query", mock.Anything, mock.Anything).Return(&dynamodb.QueryOutput{Items: []map[string]types.AttributeValue{}}, nil)

	_, err := repo.FindByID(context.Background(), "user-999")

	require.ErrorIs(t, err, domain.ErrNotFound)
	client.AssertExpectations(t)
}

func TestUserRepository_FindByID_Error(t *testing.T) {
	t.Parallel()

	client := new(mockDynamoClient)
	repo := NewUserRepository(client, "gym-tracker")

	client.On("Query", mock.Anything, mock.Anything).Return(nil, errors.New("dynamo error"))

	_, err := repo.FindByID(context.Background(), "user-123")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "dynamo query by id")
	client.AssertExpectations(t)
}

func TestUserRepository_Save_Success(t *testing.T) {
	t.Parallel()

	client := new(mockDynamoClient)
	repo := NewUserRepository(client, "gym-tracker")

	user := &domain.User{
		ID:           "user-new",
		Email:        "ivan@gym.com",
		PasswordHash: "hashedPassword",
		Name:         "Ivan",
		Active:       true,
		CreatedAt:    time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
		UpdatedAt:    time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
	}

	client.On("PutItem", mock.Anything, mock.MatchedBy(func(input *dynamodb.PutItemInput) bool {
		if *input.TableName != "gym-tracker" {
			return false
		}
		pk, _ := input.Item["PK"].(*types.AttributeValueMemberS)
		sk, _ := input.Item["SK"].(*types.AttributeValueMemberS)
		gsi1pk, _ := input.Item["GSI1PK"].(*types.AttributeValueMemberS)

		return pk != nil && pk.Value == "USER#ivan@gym.com" &&
			sk != nil && sk.Value == "PROFILE" &&
			gsi1pk != nil && gsi1pk.Value == "USER_ID#user-new" &&
			*input.ConditionExpression == "attribute_not_exists(PK)"
	})).Return(&dynamodb.PutItemOutput{}, nil)

	err := repo.Save(context.Background(), user)

	require.NoError(t, err)
	client.AssertExpectations(t)
}

func TestUserRepository_Save_AlreadyExists(t *testing.T) {
	t.Parallel()

	client := new(mockDynamoClient)
	repo := NewUserRepository(client, "gym-tracker")

	user := &domain.User{
		ID:           "user-new",
		Email:        "ivan@gym.com",
		PasswordHash: "hashedPassword",
		Name:         "Ivan",
		Active:       true,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}

	condErr := &types.ConditionalCheckFailedException{
		Message: aws.String("item already exists"),
	}

	client.On("PutItem", mock.Anything, mock.Anything).Return(nil, condErr)

	err := repo.Save(context.Background(), user)

	require.ErrorIs(t, err, domain.ErrAlreadyExists)
	client.AssertExpectations(t)
}

func TestUserRepository_Save_Error(t *testing.T) {
	t.Parallel()

	client := new(mockDynamoClient)
	repo := NewUserRepository(client, "gym-tracker")

	user := &domain.User{
		ID:           "user-new",
		Email:        "ivan@gym.com",
		PasswordHash: "hashedPassword",
		Name:         "Ivan",
		Active:       true,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}

	client.On("PutItem", mock.Anything, mock.Anything).Return(nil, errors.New("dynamo error"))

	err := repo.Save(context.Background(), user)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "dynamo put item")
	client.AssertExpectations(t)
}

func TestUserRepository_Update_Success(t *testing.T) {
	t.Parallel()

	client := new(mockDynamoClient)
	repo := NewUserRepository(client, "gym-tracker")

	user := &domain.User{
		ID:        "user-123",
		Email:     "ivan@gym.com",
		Name:      "Ivan Updated",
		Active:    false,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	client.On("UpdateItem", mock.Anything, mock.MatchedBy(func(input *dynamodb.UpdateItemInput) bool {
		pk, _ := input.Key["PK"].(*types.AttributeValueMemberS)
		sk, _ := input.Key["SK"].(*types.AttributeValueMemberS)

		return *input.TableName == "gym-tracker" &&
			pk.Value == "USER#ivan@gym.com" &&
			sk.Value == "PROFILE" &&
			*input.UpdateExpression == "SET #n = :name, active = :active, updated_at = :updatedAt" &&
			*input.ConditionExpression == "attribute_exists(PK)"
	})).Return(&dynamodb.UpdateItemOutput{}, nil)

	err := repo.Update(context.Background(), user)

	require.NoError(t, err)
	client.AssertExpectations(t)
}

func TestUserRepository_Update_NotFound(t *testing.T) {
	t.Parallel()

	client := new(mockDynamoClient)
	repo := NewUserRepository(client, "gym-tracker")

	user := &domain.User{
		ID:        "user-999",
		Email:     "nonexistent@gym.com",
		Name:      "Nobody",
		Active:    true,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	condErr := &types.ConditionalCheckFailedException{
		Message: aws.String("item does not exist"),
	}

	client.On("UpdateItem", mock.Anything, mock.Anything).Return(nil, condErr)

	err := repo.Update(context.Background(), user)

	require.ErrorIs(t, err, domain.ErrNotFound)
	client.AssertExpectations(t)
}

func TestUserRepository_Update_Error(t *testing.T) {
	t.Parallel()

	client := new(mockDynamoClient)
	repo := NewUserRepository(client, "gym-tracker")

	user := &domain.User{
		ID:        "user-123",
		Email:     "ivan@gym.com",
		Name:      "Ivan",
		Active:    true,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	client.On("UpdateItem", mock.Anything, mock.Anything).Return(nil, errors.New("dynamo error"))

	err := repo.Update(context.Background(), user)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "dynamo update item")
	client.AssertExpectations(t)
}

func TestToRecord_KeyConstruction(t *testing.T) {
	t.Parallel()

	user := &domain.User{
		ID:           "user-abc",
		Email:        "test@gym.com",
		PasswordHash: "hash123",
		Name:         "Test User",
		Active:       true,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}

	rec := toRecord(user)

	assert.Equal(t, "USER#test@gym.com", rec.PK)
	assert.Equal(t, "PROFILE", rec.SK)
	assert.Equal(t, "USER_ID#user-abc", rec.GSI1PK)
	assert.Equal(t, "PROFILE", rec.GSI1SK)
	assert.Equal(t, "user-abc", rec.ID)
	assert.Equal(t, "test@gym.com", rec.Email)
}
