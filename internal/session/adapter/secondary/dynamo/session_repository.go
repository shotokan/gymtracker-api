package dynamo

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/shotokan/gymtracker-api/internal/session/domain"
)

type client interface {
	PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
	Query(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error)
}

// DynamoDB record structs — anti-corruption layer.
type setRec struct {
	Weight      float64 `dynamodbav:"weight"`
	Reps        int     `dynamodbav:"reps"`
	RestSeconds int     `dynamodbav:"rest_seconds"`
}

type exerciseRec struct {
	ExerciseID string   `dynamodbav:"exercise_id"`
	Name       string   `dynamodbav:"name"`
	Sets       []setRec `dynamodbav:"sets"`
}

type sessionRec struct {
	PK           string        `dynamodbav:"PK"`
	SK           string        `dynamodbav:"SK"`
	GSI1PK       string        `dynamodbav:"GSI1PK"`
	GSI1SK       string        `dynamodbav:"GSI1SK"`
	ID           string        `dynamodbav:"id"`
	UserID       string        `dynamodbav:"user_id"`
	PlanID       string        `dynamodbav:"plan_id,omitempty"`
	RoutineID    string        `dynamodbav:"routine_id"`
	RoutineName  string        `dynamodbav:"routine_name"`
	Date         string        `dynamodbav:"date"`
	DurationSecs int           `dynamodbav:"duration_secs"`
	Exercises    []exerciseRec `dynamodbav:"exercises"`
	StartedAt    string        `dynamodbav:"started_at"`
	FinishedAt   string        `dynamodbav:"finished_at"`
	CreatedAt    string        `dynamodbav:"created_at"`
}

// SessionRepository implements output.SessionRepository using DynamoDB.
// Sessions are stored with date-prefixed SK for chronological ordering:
//
//	PK     = USER#<userID>
//	SK     = SESSION#<date>#<sessionID>
//	GSI1PK = SESSION#<sessionID>
//	GSI1SK = USER#<userID>
type SessionRepository struct {
	db        client
	tableName string
}

// NewSessionRepository constructs the repository.
func NewSessionRepository(db client, tableName string) *SessionRepository {
	return &SessionRepository{db: db, tableName: tableName}
}

// Save persists a new session.
func (r *SessionRepository) Save(ctx context.Context, s *domain.Session) error {
	item, err := attributevalue.MarshalMap(toRecord(s))
	if err != nil {
		return fmt.Errorf("marshaling session: %w", err)
	}

	_, err = r.db.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("dynamo put session: %w", err)
	}

	return nil
}

// FindByID retrieves a specific session by userID and sessionID.
// Uses GSI1 Query for direct lookup by sessionID.
func (r *SessionRepository) FindByID(ctx context.Context, userID, sessionID string) (*domain.Session, error) {
	out, err := r.db.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(r.tableName),
		IndexName:              aws.String("GSI1"),
		KeyConditionExpression: aws.String("GSI1PK = :pk AND GSI1SK = :sk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: gsi1PK(sessionID)},
			":sk": &types.AttributeValueMemberS{Value: gsi1SK(userID)},
		},
		Limit: aws.Int32(1),
	})
	if err != nil {
		return nil, fmt.Errorf("dynamo query session: %w", err)
	}
	if len(out.Items) == 0 {
		return nil, domain.ErrSessionNotFound
	}

	return unmarshal(out.Items[0])
}

// FindAll returns sessions for a user, newest first, up to limit.
func (r *SessionRepository) FindAll(ctx context.Context, userID string, limit int) ([]domain.Session, error) {
	out, err := r.db.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(r.tableName),
		KeyConditionExpression: aws.String("PK = :pk AND begins_with(SK, :prefix)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk":     &types.AttributeValueMemberS{Value: pk(userID)},
			":prefix": &types.AttributeValueMemberS{Value: "SESSION#"},
		},
		ScanIndexForward: aws.Bool(false), // newest first
		Limit:            aws.Int32(int32(limit)),
	})
	if err != nil {
		return nil, fmt.Errorf("dynamo query sessions: %w", err)
	}

	sessions := make([]domain.Session, 0, len(out.Items))
	for _, item := range out.Items {
		s, err := unmarshal(item)
		if err != nil {
			continue
		}
		sessions = append(sessions, *s)
	}

	return sessions, nil
}

// ── Helpers ───────────────────────────────────────────────────────────────

func pk(userID string) string { return "USER#" + userID }

func sk(date, sessionID string) string { return "SESSION#" + date + "#" + sessionID }

func gsi1PK(sessionID string) string { return "SESSION#" + sessionID }

func gsi1SK(userID string) string { return "USER#" + userID }

func toRecord(s *domain.Session) sessionRec {
	exercises := make([]exerciseRec, len(s.Exercises))
	for i, ex := range s.Exercises {
		sets := make([]setRec, len(ex.Sets))
		for j, set := range ex.Sets {
			sets[j] = setRec{
				Weight:      set.Weight,
				Reps:        set.Reps,
				RestSeconds: set.RestSeconds,
			}
		}
		exercises[i] = exerciseRec{
			ExerciseID: ex.ExerciseID,
			Name:       ex.Name,
			Sets:       sets,
		}
	}

	skVal := sk(s.Date, s.ID)
	return sessionRec{
		PK:           pk(s.UserID),
		SK:           skVal,
		GSI1PK:       gsi1PK(s.ID),
		GSI1SK:       gsi1SK(s.UserID),
		ID:           s.ID,
		UserID:       s.UserID,
		PlanID:       s.PlanID,
		RoutineID:    s.RoutineID,
		RoutineName:  s.RoutineName,
		Date:         s.Date,
		DurationSecs: s.DurationSecs,
		Exercises:    exercises,
		StartedAt:    s.StartedAt.Format(time.RFC3339),
		FinishedAt:   s.FinishedAt.Format(time.RFC3339),
		CreatedAt:    s.CreatedAt.Format(time.RFC3339),
	}
}

func unmarshal(item map[string]types.AttributeValue) (*domain.Session, error) {
	var rec sessionRec
	if err := attributevalue.UnmarshalMap(item, &rec); err != nil {
		return nil, fmt.Errorf("unmarshaling session: %w", err)
	}

	exercises := make([]domain.SessionExercise, len(rec.Exercises))
	for i, ex := range rec.Exercises {
		sets := make([]domain.CompletedSet, len(ex.Sets))
		for j, set := range ex.Sets {
			sets[j] = domain.CompletedSet{
				Weight:      set.Weight,
				Reps:        set.Reps,
				RestSeconds: set.RestSeconds,
			}
		}
		exercises[i] = domain.SessionExercise{
			ExerciseID: ex.ExerciseID,
			Name:       ex.Name,
			Sets:       sets,
		}
	}

	startedAt, _ := time.Parse(time.RFC3339, rec.StartedAt)
	finishedAt, _ := time.Parse(time.RFC3339, rec.FinishedAt)
	createdAt, _ := time.Parse(time.RFC3339, rec.CreatedAt)

	return &domain.Session{
		ID:           rec.ID,
		UserID:       rec.UserID,
		PlanID:       rec.PlanID,
		RoutineID:    rec.RoutineID,
		RoutineName:  rec.RoutineName,
		Date:         rec.Date,
		DurationSecs: rec.DurationSecs,
		Exercises:    exercises,
		StartedAt:    startedAt,
		FinishedAt:   finishedAt,
		CreatedAt:    createdAt,
	}, nil
}
