package core

import (
	"context"
	"io/ioutil"
	"os"

	"google.golang.org/grpc/metadata"
)

const TokenPath = "./token.dat"

func searchToken() (string, error) {
	log.Debug("looking for existing token")
	dat, err := ioutil.ReadFile(TokenPath)
	if os.IsNotExist(err) {
		log.Warn("token file not found; will register as new client")
		return "", nil
	} else if err != nil {
		log.Error(err)

		return "", err
	}

	log.Debug("found token")

	return string(dat), nil
}

func writeToken(token string) error {
	log.Debug("persisting token to disk")
	err := ioutil.WriteFile(TokenPath, []byte(token), 0644)
	if err != nil {
		log.Error(err)

		return err
	}

	log.Debug("done persisting token")

	return nil
}

func (c *Core) contextWithToken() context.Context {
	ctx := context.Background()

	// if we found any token persisted, use that to identify the client with the server
	if c.token != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, "token", c.token)
	}

	return ctx
}

func (c *Core) updateToken(token string) {
	if c.token == token {
		return
	}

	c.token = token
	err := writeToken(token)
	if err != nil {
		log.Fatal(err)
	}
}
