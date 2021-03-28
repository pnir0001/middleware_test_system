package main

import (
	"fmt"
	"net/http"

	"github.com/pnir0001/middleware_test_system/api/src/service"
)

func main() {

	// 各種接続
	natsConn, err := service.ConnectNats()
	if err != nil {
		fmt.Println("nats connect error")
		fmt.Println(err)
		return
	}

	redisClient := service.ConnectRedis()
	if err != nil {
		fmt.Println("redis connect error")
		fmt.Println(err)
		return
	}

	postgresDB, err := service.ConnectPostgres()
	if err != nil {
		fmt.Println("postgres connect error")
		fmt.Println(err)
		return
	}

	mongoDB, err := service.ConnectMongo()
	if err != nil {
		fmt.Println("mongo connect error")
		fmt.Println(err)
		return
	}

	s := service.Service{
		NatsConn:    natsConn,
		RedisClient: redisClient,
		Postgres:    postgresDB,
		Mongo:       mongoDB,
	}

	fmt.Println("start server")
	http.HandleFunc("/test", s.TestHandler)
	http.ListenAndServe(":8080", nil)

}