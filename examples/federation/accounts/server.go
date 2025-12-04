package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/tailor-platform/graphql"
	"github.com/tailor-platform/graphql/playground"
	"log"
	"net/http"
)

const ServerPort = 8001

type PostData struct {
	Query     string                 `json:"query"`
	Operation string                 `json:"operation"`
	Variables map[string]interface{} `json:"variables"`
}

type Service struct {
	Sdl string `json:"sdl"`
}

func main() {
	userType := graphql.NewObject(graphql.ObjectConfig{
		Name: "User",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.NewNonNull(graphql.ID),
			},
			"username": &graphql.Field{
				Type: graphql.NewNonNull(graphql.String),
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
	})

	_serviceObjectType := graphql.NewObject(graphql.ObjectConfig{
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
		Types: []*graphql.Object{userType},
		ResolveType: func(p graphql.ResolveTypeParams) *graphql.Object {
			if val, ok := p.Value.(map[string]interface{}); ok {
				if typeName, ok := val["__typename"].(string); ok {
					switch typeName {
					case "User":
						return userType
					}
				}
			}
			return nil
		},
	})

	schemaConfig := graphql.SchemaConfig{
		Query: graphql.NewObject(graphql.ObjectConfig{
			Name: "Query",
			Fields: graphql.Fields{
				"me": &graphql.Field{
					Name: "me",
					Type: userType,
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						return map[string]interface{}{
							"id":       "1234",
							"username": "Me",
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
						if id, ok := p.Args["id"].(string); ok {
							return map[string]interface{}{
								"id":       id,
								"username": "Me",
							}, nil
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
		log.Printf("accounts req=%v", p)
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
