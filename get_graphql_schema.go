package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/Yamashou/gqlgenc/client"
	"github.com/Yamashou/gqlgenc/introspection"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/formatter"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
)

func main() {
	if err := getGraphQLSchemaMain(); err != nil {
		log.Fatalf("error: %+v", err)
	}
}

func getGraphQLSchemaMain() error {
	var headerOption string
	flag.StringVar(&headerOption, "h", "", "HTTP request header 'HEADER1=Hello HEADER2=Hi'")
	flag.Parse()

	args := flag.Args()
	if len(args) != 1 {
		return fmt.Errorf("get-graphql-schema [OPTIONS] ENDPOINT: endpoint not found")
	}
	endpoint := args[0]

	header, err := parseHeaderOption(headerOption)
	if err != nil {
		return fmt.Errorf("validation: %w", err)
	}

	schemaDocument, err := getGraphQLSchemaDocument(endpoint, header)
	if err != nil {
		return fmt.Errorf("getGraphQLSchema: %w", err)
	}
	sort.Slice(schemaDocument.Directives, func(i, j int) bool { return schemaDocument.Directives[i].Name < schemaDocument.Directives[j].Name })
	sort.Slice(schemaDocument.Definitions, func(i, j int) bool { return schemaDocument.Definitions[i].Name < schemaDocument.Definitions[j].Name })
	sort.Slice(schemaDocument.Extensions, func(i, j int) bool { return schemaDocument.Extensions[i].Name < schemaDocument.Extensions[j].Name })

	astFormatter := formatter.NewFormatter(os.Stdout)
	astFormatter.FormatSchemaDocument(schemaDocument)

	return nil
}

func parseHeaderOption(headers string) (http.Header, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")
	if headers == "" {
		return header, nil
	}

	for _, h := range strings.Split(headers, ",") {
		key, value, found := strings.Cut(h, "=")
		if !found {
			return nil, fmt.Errorf("invalid header: %s", h)
		}
		header.Add(key, value)
	}
	return header, nil
}

func getGraphQLSchemaDocument(endpoint string, header http.Header) (*ast.SchemaDocument, error) {
	addHeader := func(req *http.Request) {
		req.Header = header
	}
	gqlclient := client.NewClient(http.DefaultClient, endpoint, addHeader)

	var res introspection.Query
	if err := gqlclient.Post(context.Background(), "Query", introspection.Introspection, &res, nil); err != nil {
		return nil, fmt.Errorf("introspection query failed: %w", err)
	}

	return introspection.ParseIntrospectionQuery(endpoint, res), nil
}
