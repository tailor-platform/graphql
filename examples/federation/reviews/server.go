package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/tailor-platform/graphql"
	"github.com/tailor-platform/graphql/federation/reviews/models"
	"github.com/tailor-platform/graphql/playground"
	"log"
	"net/http"
)

const ServerPort = 8003

type PostData struct {
	Query     string                 `json:"query"`
	Operation string                 `json:"operation"`
	Variables map[string]interface{} `json:"variables"`
}

type Service struct {
	Sdl string `json:"sdl"`
}

func main() {
	reviewFields := graphql.Fields{}
	reviewType := graphql.NewObject(graphql.ObjectConfig{Name: "Review", Fields: reviewFields})
	userType := graphql.NewObject(graphql.ObjectConfig{
		Name: "User",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.NewNonNull(graphql.ID),
				Directives: []*graphql.ObjectDirective{
					{
						Directive: graphql.ExternalDirective,
					},
				},
			},
			"username": &graphql.Field{
				Type: graphql.NewNonNull(graphql.String),
				Directives: []*graphql.ObjectDirective{
					{
						Directive: graphql.ExternalDirective,
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					obj := p.Source.(*models.User)
					return fmt.Sprintf("User %s", obj.ID), nil
				},
			},
			"reviews": &graphql.Field{
				Type: graphql.NewList(reviewType),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					var res []*models.Review
					obj := p.Source.(*models.User)
					for _, review := range models.Reviews {
						if review.Author.ID == obj.ID {
							res = append(res, review)
						}
					}
					return res, nil
				},
			},
		},
		Directives: []*graphql.ObjectDirective{
			{
				Directive: graphql.KeyDirective,
				Args: []graphql.ObjectDirectiveArg{
					{
						Name:  "fields",
						Value: "id",
					},
				},
			},
		},
		Extend: true,
	})

	productType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Product",
		Fields: graphql.Fields{
			"upc": &graphql.Field{
				Type: graphql.NewNonNull(graphql.String),
				Directives: []*graphql.ObjectDirective{
					{
						Directive: graphql.ExternalDirective,
					},
				},
			},
			"reviews": &graphql.Field{
				Type: graphql.NewList(reviewType),
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
		Extend: true,
	})

	reviewFields["body"] = &graphql.Field{
		Type: graphql.NewNonNull(graphql.String),
	}
	reviewFields["author"] = &graphql.Field{
		Type: graphql.NewNonNull(userType),
		Directives: []*graphql.ObjectDirective{
			{
				Directive: graphql.ProvidesDirective,
				Args: []graphql.ObjectDirectiveArg{
					{
						Name:  "fields",
						Value: "username",
					},
				},
			},
		},
	}
	reviewFields["product"] = &graphql.Field{
		Type: graphql.NewNonNull(productType),
	}

	_entityType := graphql.NewUnion(graphql.UnionConfig{
		Name:  "_Entity",
		Types: []*graphql.Object{userType, productType},
		ResolveType: func(p graphql.ResolveTypeParams) *graphql.Object {
			switch p.Value.(type) {
			case *models.User:
				return userType
			case *models.Product:
				return productType
			}
			return nil
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

	schemaConfig := graphql.SchemaConfig{
		Query: graphql.NewObject(graphql.ObjectConfig{
			Name: "Query",
			Fields: graphql.Fields{
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
									case "User":
										if id, ok := repMap["id"].(string); ok {
											resp = append(resp, &models.User{ID: id})
										}
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
			},
			Extend: true,
		}),
	}

	schema, err := graphql.NewSchema(schemaConfig)
	if err != nil {
		log.Fatalf("failed to create new schema, error: %v", err)
	}

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
