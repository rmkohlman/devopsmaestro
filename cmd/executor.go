package cmd

import (
	"context"
	"devopsmaestro/db"
	"fmt"

	"github.com/rmkohlman/MaestroSDK/render"
)

type Executor interface {
	Execute(ctx context.Context) error
}

type DefaultExecutor struct {
	DataStore db.DataStore
}

func (e *DefaultExecutor) Execute(ctx context.Context) error {
	// Access the dataStore and perform some actions
	if e.DataStore == nil {
		return fmt.Errorf("datastore is not initialized")
	}

	// Example logic: perform operations using DataStore
	render.Progress("Executing command with datastore...")
	// Insert business logic here

	return nil
}

func NewExecutor(dataStore db.DataStore) Executor {
	return &DefaultExecutor{
		DataStore: dataStore,
	}
}
