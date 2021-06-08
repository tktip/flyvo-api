package datastore

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/datastore"
	"google.golang.org/api/option"
)

// ErrNoSuchEntity indicates no entity was found on a given ID
var ErrNoSuchEntity = datastore.ErrNoSuchEntity

// Client contains config and methods to communicate with the datastore used to store refresh-tokens
type Client struct {
	CredentialsFile string `yaml:"credentialsFile"`
	GoogleProjectID string `yaml:"googleProjectID"`

	store *datastore.Client
}

// RefreshToken is a refresh token entry from the datastore
type RefreshToken struct {
	TokenString string    `datastore:"token"`
	Created     time.Time `datastore:"created"`
	Expires     time.Time `datastore:"expires"`
}

// Init initialises
func (c *Client) Init() (err error) {
	c.store, err = datastore.NewClient(
		context.TODO(),
		c.GoogleProjectID,
		option.WithCredentialsFile(c.CredentialsFile),
	)
	if err != nil {
		return fmt.Errorf("creating client: %v", err)
	}

	return nil
}

// GetRefreshToken gets the token on the given email in the datastore
func (c *Client) GetRefreshToken(ctx context.Context, userEmail string) (
	tok RefreshToken,
	err error,
) {
	err = c.store.Get(ctx, &datastore.Key{Kind: "refreshToken", Name: userEmail}, &tok)
	return
}

// PutRefreshToken puts the given token on the given email in the datastore
func (c *Client) PutRefreshToken(ctx context.Context, userEmail string, tok RefreshToken) (
	err error,
) {
	_, err = c.store.Put(ctx, &datastore.Key{Kind: "refreshToken", Name: userEmail}, &tok)
	return
}

// PopRefreshToken deletes the refresh-token by the given email
func (c *Client) PopRefreshToken(ctx context.Context, userEmail string) (
	err error,
) {
	err = c.store.Delete(ctx, &datastore.Key{Kind: "refreshToken", Name: userEmail})
	return
}

// Close closes the datastore client's connection
func (c *Client) Close() error {
	return c.store.Close()
}
