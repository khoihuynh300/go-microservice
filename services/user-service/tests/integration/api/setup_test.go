package api_test

import (
	"context"
	"log"
	"os"
	"testing"

	userpb "github.com/khoihuynh300/go-microservice/shared/proto/user"
	"github.com/khoihuynh300/go-microservice/user-service/tests/integration/testutil"
)

var (
	testDB     *testutil.TestDatabase
	testRedis  *testutil.TestRedis
	testServer *testutil.TestGRPCServer
	client     userpb.UserServiceClient
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	var err error
	testDB, err = testutil.StartPostgresContainer(ctx)
	if err != nil {
		log.Fatalf("Failed to start postgres container: %v", err)
	}

	testRedis, err = testutil.StartRedisContainer(ctx)
	if err != nil {
		log.Fatalf("Failed to start redis container: %v", err)
	}

	testServer, err = testutil.NewTestGRPCServer(ctx, testDB, testRedis, nil)
	if err != nil {
		log.Fatalf("Failed to start grpc server: %v", err)
	}

	client = testServer.GetClient()

	code := m.Run()

	testServer.TearDown()
	testRedis.TearDown(ctx)
	testDB.TearDown(ctx)

	os.Exit(code)
}

// cleanupTestData clears all test data between tests
func cleanupTestData(ctx context.Context) error {
	if err := testDB.CleanupTestData(ctx); err != nil {
		return err
	}
	return testRedis.FlushAll(ctx)
}
