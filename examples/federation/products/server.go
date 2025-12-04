package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/tailor-platform/graphql"
	"github.com/tailor-platform/graphql/federation/products/models"
	"github.com/tailor-platform/graphql/playground"
	"log"
	"math/rand"
	"net/http"
	"time"
)

const ServerPort = 8002

type PostData struct {
	Query     string                 `json:"query"`
	Operation string                 `json:"operation"`
	Variables map[string]interface{} `json:"variables"`
}

type Service struct {
	Sdl string `json:"sdl"`
}

var (
	randomnessEnabled = true
	minPrice          = 10
	maxPrice          = 1499
	currentPrice      = minPrice
	updateInterval    = time.Second
)

func main() {

	productType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Product",
		Fields: graphql.Fields{
			"upc": &graphql.Field{
				Type: graphql.NewNonNull(graphql.String),
			},
			"name": &graphql.Field{
				Type: graphql.NewNonNull(graphql.String),
			},
			"price": &graphql.Field{
				Type: graphql.NewNonNull(graphql.Int),
			},
			"inStock": &graphql.Field{
				Type: graphql.NewNonNull(graphql.Int),
			},
		},
		Directives: []*graphql.ObjectDirective{
			{
				Directive: graphql.KeyDirective,
				Args: []graphql.ObjectDirectiveArg{
					{
						Name:  "fields",
						Value: "upc",
					},
				},
			},
		},
	})

	var _serviceObjectType = graphql.NewObject(graphql.ObjectConfig{
		Name: "_Service",
		Fields: graphql.Fields{
			"sdl": &graphql.Field{
				Name: "sdl",
				Type: graphql.String,
			},
		},
	})

	_entityType := graphql.NewUnion(graphql.UnionConfig{
		Name:  "_Entity",
		Types: []*graphql.Object{productType},
		ResolveType: func(p graphql.ResolveTypeParams) *graphql.Object {
			switch p.Value.(type) {
			case *models.Product:
				return productType
			}
			return nil
		},
	})

	schemaConfig := graphql.SchemaConfig{
		Query: graphql.NewObject(graphql.ObjectConfig{
			Name: "Query",
			Fields: graphql.Fields{
				"topProducts": &graphql.Field{
					Name: "topProducts",
					Type: graphql.NewList(productType),
					Args: graphql.FieldConfigArgument{
						"first": &graphql.ArgumentConfig{
							Type:         graphql.Int,
							DefaultValue: 5,
						},
					},
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						return models.Hats, nil
					},
				},
				"_entities": &graphql.Field{
					Name: "_entities",
					Type: &graphql.NonNull{
						OfType: graphql.NewList(_entityType),
					},
					Args: graphql.FieldConfigArgument{
						"representations": &graphql.ArgumentConfig{
							Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(graphql.Any))),
						},
					},
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						if reps, ok := p.Args["representations"].([]interface{}); ok {
							var resp []interface{}
							for _, rep := range reps {
								repMap := rep.(map[string]interface{})
								if typename, ok := repMap["__typename"].(string); ok {
									switch typename {
									case "Product":
										if upc, ok := repMap["upc"].(string); ok {
											resp = append(resp, &models.Product{Upc: upc})
										}
									}
								}
							}
							return resp, nil
						}
						return nil, nil
					},
				},
				"_service": &graphql.Field{
					Name: "_service",
					Type: graphql.NewNonNull(_serviceObjectType),
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						sdl := graphql.BuildSDL(p.Info.Schema, &graphql.SDLExportOptions{
							HideDoubleUnderscorePrefix: true,
							IncludeBasicScalar:         false,
						})
						return Service{
							Sdl: sdl,
						}, nil
					},
				},
			},
			Extend: true,
		}),
		Subscription: graphql.NewObject(graphql.ObjectConfig{
			Name: "Subscription",
			Fields: graphql.Fields{
				"updatedPrice": &graphql.Field{
					Type: graphql.NewNonNull(productType),
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						ctx := p.Context
						updatedPrice := make(chan *models.Product)
						go func() {
							for {
								select {
								case <-ctx.Done():
									return
								case <-time.After(updateInterval):
									rand.Seed(time.Now().UnixNano())
									product := models.Hats[0]

									if randomnessEnabled {
										product = models.Hats[rand.Intn(len(models.Hats)-1)]
										product.Price = rand.Intn(maxPrice-minPrice+1) + minPrice
										updatedPrice <- product
										continue
									}

									product.Price = currentPrice
									currentPrice += 1
									updatedPrice <- product
								}
							}
						}()
						return updatedPrice, nil
					},
				},
				"updateProductPrice": &graphql.Field{
					Type: productType,
					Args: graphql.FieldConfigArgument{
						"upc": &graphql.ArgumentConfig{
							Type:         graphql.NewNonNull(graphql.String),
							DefaultValue: 5,
						},
					},
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						ctx := p.Context
						upc := p.Args["upc"].(string)
						updatedPrice := make(chan *models.Product)
						var product *models.Product

						for _, hat := range models.Hats {
							if hat.Upc == upc {
								product = hat
								break
							}
						}

						if product == nil {
							return nil, fmt.Errorf("unknown product upc: %s", upc)
						}

						go func() {
							for {
								select {
								case <-ctx.Done():
									return
								case <-time.After(time.Second):
									rand.Seed(time.Now().UnixNano())
									min := 10
									max := 1499
									product.Price = rand.Intn(max-min+1) + min
									updatedPrice <- product
								}
							}
						}()
						return updatedPrice, nil
					},
				},
				"stock": &graphql.Field{
					Type: graphql.NewList(graphql.NewNonNull(productType)),
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						ctx := p.Context
						stock := make(chan []*models.Product)

						go func() {
							for {
								select {
								case <-ctx.Done():
									return
								case <-time.After(2 * time.Second):
									rand.Seed(time.Now().UnixNano())
									randIndex := rand.Intn(len(models.Hats))

									if models.Hats[randIndex].InStock > 0 {
										models.Hats[randIndex].InStock--
									}

									stock <- models.Hats
								}
							}
						}()

						return stock, nil
					},
				},
			},
			Extend: true,
		}),
	}

	schema, err := graphql.NewSchema(schemaConfig)
	if err != nil {
		log.Fatalf("failed to create new schema, error: %v", err)
	}

	log.Printf("%v", graphql.BuildSDL(schema, &graphql.SDLExportOptions{
		HideDoubleUnderscorePrefix: true,
		IncludeBasicScalar:         false,
	}))
	r := mux.NewRouter()
	r.HandleFunc("/graphql", func(writer http.ResponseWriter, request *http.Request) {
		ctx := request.Context()
		var p PostData
		if err := json.NewDecoder(request.Body).Decode(&p); err != nil {
			writer.WriteHeader(400)
			return
		}

		result := graphql.Do(graphql.Params{
			Schema:         schema,
			RequestString:  p.Query,
			VariableValues: p.Variables,
			OperationName:  p.Operation,
			Context:        ctx,
		})
		if err := json.NewEncoder(writer).Encode(result); err != nil {
			log.Printf("could not write result to response: %s", err)
			writer.WriteHeader(503)
		}
	})

	r.HandleFunc("/playground", playground.Handler("playground", "/graphql"))

	http.Handle("/", r)
	log.Printf("Now server is running on port %d\n", ServerPort)
	log.Printf("Playground: 'http://localhost:%d/playground'\n", ServerPort)
	http.ListenAndServe(fmt.Sprintf(":%d", ServerPort), nil)
}
