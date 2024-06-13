package example

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/cucumber/godog"
	"github.com/jfelipearaujo/testcontainers/pkg/container"
	"github.com/jfelipearaujo/testcontainers/pkg/container/mongodb"
	"github.com/jfelipearaujo/testcontainers/pkg/state"
	"github.com/jfelipearaujo/testcontainers/pkg/testsuite"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	ErrProductNotFound = errors.New("product not found")
)

type product struct {
	Id   primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name string             `json:"name" bson:"name"`
	Desc string             `json:"description" bson:"description"`
}

type test struct {
	connStr string

	productId   primitive.ObjectID
	productName string
	productDesc string
}

var testState = state.NewState[test]()

var containers = container.NewGroup()

func TestFeatures(t *testing.T) {
	testsuite.NewTestSuite(t,
		initializeScenario,
		testsuite.WithPaths("features"),
		testsuite.WithConcurrency(0),
	)
}

func initializeScenario(ctx *godog.ScenarioContext) {
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		definition := container.NewContainerDefinition(
			mongodb.WithMongoContainer(),
		)

		mongoContainer, err := definition.BuildContainer(ctx)
		if err != nil {
			return ctx, err
		}

		connectionString, err := mongodb.BuildInternalConnectionString(ctx, mongoContainer)
		if err != nil {
			return ctx, err
		}

		currentState := testState.Retrieve(ctx)
		currentState.connStr = connectionString

		containers[sc.Id] = container.BuildGroupContainer(
			container.WithDockerContainer(mongoContainer),
		)

		return testState.Enrich(ctx, currentState), nil
	})

	ctx.Step(`^I have a product with name "([^"]*)"$`, iHaveAProductWithName)
	ctx.Step(`^I have a product with description "([^"]*)"$`, iHaveAProductWithDescription)
	ctx.Step(`^I create the product$`, iCreateTheProduct)
	ctx.Step(`^I update the product$`, iUpdateTheProduct)
	ctx.Step(`^I delete the product$`, iDeleteTheProduct)
	ctx.Step(`^I update the product name to "([^"]*)"$`, iUpdateTheProductNameTo)
	ctx.Step(`^the product should be created$`, theProductShouldBeCreated)
	ctx.Step(`^the product should be updated$`, theProductShouldBeUpdated)
	ctx.Step(`^the product should be deleted$`, theProductShouldBeDeleted)

	ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		group := containers[sc.Id]

		return container.DestroyGroup(ctx, group)
	})
}

func iHaveAProductWithName(ctx context.Context, productName string) (context.Context, error) {
	currentState := testState.Retrieve(ctx)
	currentState.productName = productName
	return testState.Enrich(ctx, currentState), nil
}

func iHaveAProductWithDescription(ctx context.Context, productDescription string) (context.Context, error) {
	currentState := testState.Retrieve(ctx)
	currentState.productDesc = productDescription
	return testState.Enrich(ctx, currentState), nil
}

func iUpdateTheProductNameTo(ctx context.Context, newProductName string) (context.Context, error) {
	currentState := testState.Retrieve(ctx)
	currentState.productName = newProductName
	return testState.Enrich(ctx, currentState), nil
}

func iCreateTheProduct(ctx context.Context) (context.Context, error) {
	currentState := testState.Retrieve(ctx)

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(currentState.connStr))
	if err != nil {
		return ctx, fmt.Errorf("failed to connect to the database: %w", err)
	}
	defer client.Disconnect(ctx)

	result, err := client.Database("test").Collection("products").InsertOne(ctx, product{
		Name: currentState.productName,
		Desc: currentState.productDesc,
	})
	if err != nil {
		return ctx, fmt.Errorf("failed to insert product: %w", err)
	}

	currentState.productId = result.InsertedID.(primitive.ObjectID)

	return testState.Enrich(ctx, currentState), nil
}

func iUpdateTheProduct(ctx context.Context) (context.Context, error) {
	currentState := testState.Retrieve(ctx)

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(currentState.connStr))
	if err != nil {
		return ctx, fmt.Errorf("failed to connect to the database: %w", err)
	}
	defer client.Disconnect(ctx)

	var product product

	if err := client.Database("test").Collection("products").FindOneAndUpdate(
		ctx,
		bson.M{"_id": currentState.productId},
		bson.M{"$set": bson.M{"name": currentState.productName}},
		options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(&product); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return ctx, ErrProductNotFound
		}
		return ctx, fmt.Errorf("failed to update product: %w", err)
	}

	return ctx, nil
}

func iDeleteTheProduct(ctx context.Context) (context.Context, error) {
	currentState := testState.Retrieve(ctx)

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(currentState.connStr))
	if err != nil {
		return ctx, fmt.Errorf("failed to connect to the database: %w", err)
	}
	defer client.Disconnect(ctx)

	result, err := client.Database("test").Collection("products").DeleteOne(
		ctx,
		bson.M{"_id": currentState.productId},
	)
	if err != nil {
		return ctx, fmt.Errorf("failed to delete product: %w", err)
	}

	if result.DeletedCount == 0 {
		return ctx, ErrProductNotFound
	}

	return ctx, nil
}

func theProductShouldBeCreated(ctx context.Context) (context.Context, error) {
	currentState := testState.Retrieve(ctx)

	product, err := findProductById(ctx, currentState.connStr, currentState.productId)
	if err != nil {
		return ctx, err
	}

	if product.Id != currentState.productId {
		return ctx, fmt.Errorf("product not created")
	}

	return ctx, nil
}

func theProductShouldBeUpdated(ctx context.Context) (context.Context, error) {
	currentState := testState.Retrieve(ctx)

	product, err := findProductById(ctx, currentState.connStr, currentState.productId)
	if err != nil {
		return ctx, err
	}

	if product.Name != currentState.productName {
		return ctx, fmt.Errorf("product not updated")
	}

	return ctx, nil
}

func theProductShouldBeDeleted(ctx context.Context) (context.Context, error) {
	currentState := testState.Retrieve(ctx)

	_, err := findProductById(ctx, currentState.connStr, currentState.productId)
	if err != ErrProductNotFound {
		return ctx, fmt.Errorf("product not deleted: %w", err)
	}

	return ctx, nil
}

func findProductById(ctx context.Context, connStr string, productId primitive.ObjectID) (product, error) {
	var product product

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(connStr))
	if err != nil {
		return product, fmt.Errorf("failed to connect to the database: %w", err)
	}
	defer client.Disconnect(ctx)

	result := client.Database("test").Collection("products").FindOne(ctx, bson.M{"_id": productId})

	if err = result.Decode(&product); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return product, ErrProductNotFound
		}
		return product, fmt.Errorf("failed to decode product: %w", err)
	}
	return product, nil
}
