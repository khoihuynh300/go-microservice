package repository_test

import (
	"context"
	"os"
	"testing"

	"github.com/khoihuynh300/go-microservice/user-service/tests/integration/testutil"
)

var testDB *testutil.TestDatabase

func TestMain(m *testing.M) {
	ctx := context.Background()

	var err error
	testDB, err = testutil.StartPostgresContainer(ctx)
	if err != nil {
		panic("Failed to start test database: " + err.Error())
	}

	code := m.Run()

	testDB.TearDown(ctx)

	os.Exit(code)
}
