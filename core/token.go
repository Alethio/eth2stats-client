package core

import (
	"context"
	"io/ioutil"
	"os"
	"path"

	"google.golang.org/grpc/metadata"
)

const TokenFile = "token.dat"

func  (c *Core) searchToken() error {
	log.Debug("looking for existing token")
	fileName := path.Join(c.config.DataFolder, TokenFile)
	dat, err := ioutil.ReadFile(fileName)
	if os.IsNotExist(err) {
		log.Warn("token file not found; will register as new client")
		return nil
	} else if err != nil {
		log.Error(err)

		return err
	}

	log.Debug("found token")
	c.token = string(dat)
	return nil
}

func  (c *Core) writeToken(token string) error {
	fileName := path.Join(c.config.DataFolder, TokenFile)
	log.Debug("persisting token to disk")
	_ = os.MkdirAll(c.config.DataFolder, os.ModePerm)
	err := ioutil.WriteFile(fileName, []byte(token), 0644)
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
	err := c.writeToken(token)
	if err != nil {
		log.Fatal(err)
	}
}
