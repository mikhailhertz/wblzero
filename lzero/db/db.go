package db

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"wb/tasks/lzero/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	cmap "github.com/orcaman/concurrent-map/v2"
)

// Get/Store are probably safe for concurrency
type Db struct {
	cache      cmap.ConcurrentMap[models.Order]
	pool       *pgxpool.Pool
	stopSignal chan struct{}
}

// Creates a pool of database connections and an in-memory cache.
//
// Fills the cache with all the rows read from the database.
//
// Creates a goroutine that listens for incoming messages and stores them in cache and in the database.
func Create(url string, messageChannel <-chan string) (*Db, error) {
	cache := cmap.New[models.Order]()

	pool, err := pgxpool.New(context.Background(), url)
	if err != nil {
		return nil, err
	}

	// This reads all the rows
	rows, err := pool.Query(context.Background(), "SELECT order_uid, j FROM orders")
	if err != nil {
		return nil, err
	}

	// This fills the cache
	var orderUid, orderJson string
	_, err = pgx.ForEachRow(rows, []any{&orderUid, &orderJson}, func() error {
		var order models.Order
		err := json.Unmarshal([]byte(orderJson), &order)

		if err == nil {
			cache.Set(orderUid, order)
			return nil
		} else {
			return err
		}
	})
	if err != nil {
		return nil, err
	}

	stopSignal := make(chan struct{})
	db := Db{cache, pool, stopSignal}
	go db.Listen(messageChannel)
	return &db, nil
}

func (db *Db) Destroy() {
	db.stopSignal <- struct{}{}
	<-db.stopSignal
	db.pool.Close()
}

func (db *Db) Listen(messageChannel <-chan string) {
L:
	for {
		select {
		case message, ok := <-messageChannel:
			if ok {
				go db.Store(message)
			} else {
				fmt.Println("Message channel is closed, stopped listening")
				<-db.stopSignal
				break L
			}
		case <-db.stopSignal:
			break L
		}
	}
	db.stopSignal <- struct{}{}
	// TODO: save all the unread messages
}

func (db *Db) Get(orderUid string) (string, error) {
	// Try reading from cache
	order, ok := db.cache.Get(orderUid)
	if ok {
		jsonOrder, _ := json.MarshalIndent(&order, "", "  ")
		return string(jsonOrder), nil
	}

	// Try reading from the database
	var jsonOrder string
	row := db.pool.QueryRow(context.Background(), "SELECT jsonb_pretty(j) FROM orders WHERE order_uid=$1", orderUid)
	err := row.Scan(&jsonOrder)
	if err == nil {
		return jsonOrder, nil
	}

	// Give up
	return "", errors.New("Order not found")
}

func (db *Db) Store(message string) {
	var order models.Order
	err := json.Unmarshal([]byte(message), &order)
	if err != nil {
		fmt.Println(err)
		return
	}

	_, err = db.pool.Exec(context.Background(), `INSERT INTO orders (order_uid, j) VALUES ($1, $2)`, order.OrderUid, message)
	if err != nil {
		fmt.Println(err)
		return
	}

	db.cache.Set(order.OrderUid, order)
	fmt.Printf("Stored order with uid=%s in db and cache\n", order.OrderUid)
}
