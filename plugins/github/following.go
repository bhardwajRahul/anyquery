package main

import (
	"context"
	"fmt"
	"log"

	"github.com/dgraph-io/badger/v4"
	"github.com/google/go-github/v63/github"
	"github.com/julien040/anyquery/rpc"
)

// A constructor to create a new table instance
// This function is called everytime a new connection is made to the plugin
//
// It should return a new table instance, the database schema and if there is an error
func followingCreator(args rpc.TableCreatorArgs) (rpc.Table, *rpc.DatabaseSchema, error) {
	// Request a connection
	client, token, err := getClient(args)
	if err != nil {
		return nil, nil, err
	}

	// Open the database
	db, err := openDatabase("following", token)
	if err != nil {
		return nil, nil, err
	}
	return &followingTable{client, db}, &rpc.DatabaseSchema{
		PrimaryKey: 1,
		Columns: []rpc.DatabaseSchemaColumn{
			{
				Name:        "user",
				Type:        rpc.ColumnTypeString,
				IsParameter: true,
				IsRequired:  true,
				Description: "The user to get the following from",
			},
			{
				Name:        "follower",
				Type:        rpc.ColumnTypeString,
				Description: "The username of the follower",
			},
			{
				Name:        "follower_url",
				Type:        rpc.ColumnTypeString,
				Description: "The profile URL of the follower",
			},
		},
	}, nil
}

type followingTable struct {
	client *github.Client
	db     *badger.DB
}

type followingCursor struct {
	client *github.Client
	db     *badger.DB
	pageID int
}

// Return a slice of rows that will be returned to Anyquery and filtered.
// The second return value is true if the cursor has no more rows to return
//
// The constraints are used for optimization purposes to "pre-filter" the rows
// If the rows returned don't match the constraints, it's not an issue. Anyquery will filter them out
func (t *followingCursor) Query(constraints rpc.QueryConstraint) ([][]interface{}, bool, error) {
	// Get the user from the constraints
	user := retrieveArgString(constraints, 0)

	if user == "" {
		return nil, true, fmt.Errorf("missing user")
	}

	cacheKey := fmt.Sprintf("following-%s-%d", user, t.pageID)

	// Check the cache
	rows := [][]interface{}{}

	err := loadCache(t.db, cacheKey, &rows)
	if err == nil {
		t.pageID++
		return rows, len(rows) == 0, nil
	}

	// Get the followers from the API
	followers, _, err := t.client.Users.ListFollowing(context.Background(), user, &github.ListOptions{
		Page:    t.pageID,
		PerPage: 100,
	})

	if err != nil {
		return nil, true, fmt.Errorf("failed to get followers: %w", err)
	}

	for _, follower := range followers {
		rows = append(rows, []interface{}{
			follower.GetLogin(),
			follower.GetHTMLURL(),
		})
	}

	// Save the rows in the cache
	err = saveCache(t.db, cacheKey, rows)
	if err != nil {
		log.Printf("failed to save cache: %v", err)
	}

	t.pageID++

	return rows, len(rows) == 0, nil
}

// Create a new cursor that will be used to read rows
func (t *followingTable) CreateReader() rpc.ReaderInterface {
	return &followingCursor{
		client: t.client,
		db:     t.db,
		pageID: 1,
	}
}

// A destructor to clean up resources
func (t *followingTable) Close() error {
	return nil
}
