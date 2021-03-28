package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"

	"github.com/go-redis/redis"
	nats "github.com/nats-io/go-nats-streaming"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Service Service
type Service struct {
	NatsConn    nats.Conn
	RedisClient *redis.Client
	Postgres    *sql.DB
	Mongo       *mongo.Client
}

// Response
type Response struct {
	ID   string        `json:"id"`
	Type []ReponseType `json:"type,omitempty"`
}
type ReponseType struct {
	Name      string `json:"name"`
	Timestamp int64  `json:"timestamp"`
}

// Nats

type NatsMessageStruct struct {
	ID        string `json:"id"`
	Timestamp int64  `json:"timestamp"`
}

// ConnectNats ConnectNats
func ConnectNats() (nats.Conn, error) {
	randID := uuid.New().String()
	sc, err := nats.Connect("test-cluster", randID, nats.NatsURL("test_nats_streaming:4222"))
	return sc, err
}

// PublishMessage PublishMessage
func (s *Service) PublishMessage(id string, timestamp int64) error {
	nms := NatsMessageStruct{
		ID:        id,
		Timestamp: timestamp,
	}
	bytes, _ := json.Marshal(&nms)
	return s.NatsConn.Publish("test-subject", bytes)
}

// SubscribeMessage SubscribeMessage
func (s *Service) SubscribeMessage(id string) error {
	_, err := s.NatsConn.Subscribe("test-subject", func(msg *nats.Msg) {
		s := NatsMessageStruct{}
		json.Unmarshal(msg.Data, &s)
	})
	return err
}

// Redis

// ConnectRedis ConnectRedis
func ConnectRedis() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr: "test-redis:6379",
	})

	return rdb
}

// SetRedis SetRedis
func (s *Service) SetRedis(id string, timestamp int64) error {
	strTimestamp := strconv.Itoa(int(timestamp))
	err := s.RedisClient.Set(id, strTimestamp, 0).Err()
	return err
}

// GetRedis GetRedis
func (s *Service) GetRedis(id string) (int64, error) {
	val, err := s.RedisClient.Get(id).Result()
	if err != nil {
		return 0, err
	}
	timestamp, _ := strconv.Atoi(val)
	return int64(timestamp), nil
}

// Postgres

// ConnectPostgres ConnectPostgres
func ConnectPostgres() (*sql.DB, error) {
	return sql.Open("postgres", "host=test_postgres port=5432 user=user1 password=password dbname=test_postgres_db sslmode=disable")
}

// InsertPostgres InsertPostgres
func (s *Service) InsertPostgres(id string, timestamp int64) error {
	_, err := s.Postgres.Query("INSERT INTO ids(id, timestamp) VALUES($1,$2)", id, timestamp)
	return err
}

// SelectPostgres SelectPostgres
func (s *Service) SelectPostgres(id string) (int64, error) {
	rows, err := s.Postgres.Query("SELECT timestamp FROM ids WHERE id = $1", id)
	if err != nil {
		return 0, nil
	}

	var timestamp int64
	for rows.Next() {
		rows.Scan(&timestamp)
	}

	if timestamp == 0 {
		err := fmt.Errorf("Postgres record not found.")
		return 0, err
	}

	return timestamp, nil
}

// Mongo

// ConnectMongo ConnectMongo
func ConnectMongo() (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return mongo.Connect(ctx, options.Client().ApplyURI("mongodb://test_mongo:27017"))
}

// InsertMongo InsertMongo
func (s *Service) InsertMongo(id string, timestamp int64) error {
	collection := s.Mongo.Database("test_mongo_db").Collection("requests")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := collection.InsertOne(ctx, bson.D{{"id", id}, {"timestamp", timestamp}})
	return err
}

// FindMongo FindMongo
func (s *Service) FindMongo(id string) (int64, error) {
	collection := s.Mongo.Database("test_mongo_db").Collection("requests")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cur, err := collection.Find(ctx, bson.D{{"id", id}})
	if err != nil {
		return 0, err
	}
	defer cur.Close(ctx)

	var timestamp int64
	for cur.Next(ctx) {
		var result bson.D
		err := cur.Decode(&result)
		if err != nil {
			return 0, err
		}
		for _, item := range result {
			if "timestamp" == item.Key {
				timestamp = item.Value.(int64)
			}
		}
	}
	if err := cur.Err(); err != nil {
		return 0, err
	}
	if timestamp == 0 {
		err := fmt.Errorf("Mongo record not found.")
		return 0, err
	}
	return timestamp, nil
}

// Hnadler

// TestHandler TestHandler
func (s *Service) TestHandler(w http.ResponseWriter, r *http.Request) {
	var id string
	var rt []ReponseType
	if id = r.URL.Query().Get("id"); id != "" {
		rt = []ReponseType{}

		redisTimestamp, err := s.GetRedis(id)
		if err != nil {
			fmt.Println(err)
		}
		rt = append(rt, ReponseType{Name: "redis", Timestamp: redisTimestamp})

		postgresTimestamp, err := s.SelectPostgres(id)
		if err != nil {
			fmt.Println(err)
		}
		rt = append(rt, ReponseType{Name: "postgres", Timestamp: postgresTimestamp})

		mongoTimestamp, err := s.FindMongo(id)
		if err != nil {
			fmt.Println(err)
		}
		rt = append(rt, ReponseType{Name: "mongo", Timestamp: mongoTimestamp})

	} else {
		id = uuid.New().String()
		timestamp := time.Now().UnixNano()
		err := s.PublishMessage(id, timestamp)
		if err != nil {
			fmt.Println(err)
		}
	}

	resp := Response{
		ID:   id,
		Type: rt,
	}

	b, _ := json.Marshal(resp)

	fmt.Fprintf(w, string(b))
}
