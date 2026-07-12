package main

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"

	"git.subcult.tv/subculture-collective/clpr/config"
	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
)

var ginParameter = regexp.MustCompile(`[:*]([A-Za-z0-9_]+)`)
var openAPIParameter = regexp.MustCompile(`\{([A-Za-z0-9_]+)\}`)
var operationIDPart = regexp.MustCompile(`[^A-Za-z0-9]+`)

type routeContractManifestEntry struct {
	Method        string `json:"method"`
	Path          string `json:"path"`
	Handler       string `json:"handler"`
	OperationID   string `json:"operation_id"`
	ContractLevel string `json:"contract_level"`
}

func zeroHandlers() *Handlers {
	handlers := &Handlers{}
	value := reflect.ValueOf(handlers).Elem()
	for index := 0; index < value.NumField(); index++ {
		field := value.Field(index)
		if field.CanSet() && field.Kind() == reflect.Pointer {
			field.Set(reflect.New(field.Type().Elem()))
		}
	}
	return handlers
}

func supportedRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	v1 := router.Group("/api/v1")
	cfg := &config.Config{}
	infra := &Infrastructure{Config: cfg}
	handlers := zeroHandlers()
	services := &Services{}
	registerPublicRoutes(router, v1, handlers, services, infra, cfg)
	registerAuthRoutes(v1, handlers, services, infra)
	registerClipRoutes(v1, handlers, services, infra)
	registerContentRoutes(v1, handlers, services, infra)
	registerUserRoutes(v1, handlers, services, infra)
	registerSocialRoutes(v1, handlers, services, infra)
	registerPlatformRoutes(v1, handlers, services, infra)
	registerAdminRoutes(v1, handlers, services, infra)
	return router
}

func TestSupportedRoutesHaveOpenAPIOperations(t *testing.T) {
	const filename = "../../../docs/openapi/openapi.yaml"
	if os.Getenv("UPDATE_OPENAPI_ROUTES") == "1" {
		if err := removeGeneratedOperations(filename); err != nil {
			t.Fatal(err)
		}
	}
	contents, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	var document struct {
		Paths map[string]map[string]any `yaml:"paths"`
	}
	if err := yaml.Unmarshal(contents, &document); err != nil {
		t.Fatal(err)
	}

	operations := make(map[string]struct{})
	for path, methods := range document.Paths {
		for method := range methods {
			switch strings.ToUpper(method) {
			case "GET", "POST", "PUT", "PATCH", "DELETE":
				operations[strings.ToUpper(method)+" "+path] = struct{}{}
			}
		}
	}

	var missing []string
	for _, route := range supportedRouter().Routes() {
		path := ginParameter.ReplaceAllString(route.Path, `{$1}`)
		if path != "/" {
			path = strings.TrimSuffix(path, "/")
		}
		key := route.Method + " " + path
		if _, documented := operations[key]; !documented {
			missing = append(missing, key)
		}
	}
	sort.Strings(missing)
	if len(missing) > 0 {
		if os.Getenv("UPDATE_OPENAPI_ROUTES") == "1" {
			if err := appendGeneratedOperations(filename, missing); err != nil {
				t.Fatal(err)
			}
			t.Fatalf("added %d generated operations; rerun the test", len(missing))
		}
		t.Fatalf("%d supported routes lack OpenAPI operations:\n%s", len(missing), strings.Join(missing, "\n"))
	}
}

func TestOpenAPIOperationsHaveResponseSchemas(t *testing.T) {
	contents, err := os.ReadFile("../../../docs/openapi/openapi.yaml")
	if err != nil {
		t.Fatal(err)
	}
	var document struct {
		Paths map[string]map[string]map[string]any `yaml:"paths"`
	}
	if err := yaml.Unmarshal(contents, &document); err != nil {
		t.Fatal(err)
	}

	var violations []string
	for path, pathItem := range document.Paths {
		for method, operation := range pathItem {
			switch strings.ToUpper(method) {
			case "GET", "POST", "PUT", "PATCH", "DELETE":
			default:
				continue
			}
			responses, ok := operation["responses"].(map[string]any)
			if !ok || len(responses) == 0 {
				violations = append(violations, strings.ToUpper(method)+" "+path+" has no responses")
				continue
			}
			for status, rawResponse := range responses {
				response, ok := rawResponse.(map[string]any)
				if !ok {
					violations = append(violations, strings.ToUpper(method)+" "+path+" "+status+" is malformed")
					continue
				}
				if _, referenced := response["$ref"]; referenced || status == "204" || strings.HasPrefix(status, "3") {
					continue
				}
				content, ok := response["content"].(map[string]any)
				if !ok || len(content) == 0 {
					violations = append(violations, strings.ToUpper(method)+" "+path+" "+status+" has no response schema")
					continue
				}
				for mediaType, rawMedia := range content {
					media, ok := rawMedia.(map[string]any)
					if !ok {
						violations = append(violations, strings.ToUpper(method)+" "+path+" "+status+" "+mediaType+" is malformed")
						continue
					}
					if _, hasSchema := media["schema"]; !hasSchema {
						violations = append(violations, strings.ToUpper(method)+" "+path+" "+status+" "+mediaType+" has no schema")
					}
				}
			}
		}
	}
	sort.Strings(violations)
	if len(violations) > 0 {
		t.Fatalf("%d OpenAPI responses lack schemas:\n%s", len(violations), strings.Join(violations, "\n"))
	}
}

func TestOpenAPIRouteContractManifestIsCurrent(t *testing.T) {
	const specFilename = "../../../docs/openapi/openapi.yaml"
	const manifestFilename = "../../../docs/openapi/route-contract-manifest.json"
	contents, err := os.ReadFile(specFilename)
	if err != nil {
		t.Fatal(err)
	}
	var document struct {
		Paths map[string]map[string]map[string]any `yaml:"paths"`
	}
	if err := yaml.Unmarshal(contents, &document); err != nil {
		t.Fatal(err)
	}

	entries := make([]routeContractManifestEntry, 0, len(supportedRouter().Routes()))
	for _, route := range supportedRouter().Routes() {
		path := ginParameter.ReplaceAllString(route.Path, `{$1}`)
		if path != "/" {
			path = strings.TrimSuffix(path, "/")
		}
		operation := document.Paths[path][strings.ToLower(route.Method)]
		operationID, _ := operation["operationId"].(string)
		level := "route-specific"
		if derived, _ := operation["x-clpr-router-derived"].(bool); derived {
			level = "transitional"
		}
		entries = append(entries, routeContractManifestEntry{
			Method: route.Method, Path: path, Handler: route.Handler,
			OperationID: operationID, ContractLevel: level,
		})
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Path == entries[j].Path {
			return entries[i].Method < entries[j].Method
		}
		return entries[i].Path < entries[j].Path
	})
	generated, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	generated = append(generated, '\n')
	if os.Getenv("UPDATE_OPENAPI_MANIFEST") == "1" {
		if err := os.WriteFile(manifestFilename, generated, 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	existing, err := os.ReadFile(manifestFilename)
	if err != nil {
		t.Fatalf("route contract manifest missing; run UPDATE_OPENAPI_MANIFEST=1 go test ./cmd/api -run TestOpenAPIRouteContractManifestIsCurrent: %v", err)
	}
	if string(existing) != string(generated) {
		t.Fatal("route contract manifest is stale; regenerate with UPDATE_OPENAPI_MANIFEST=1")
	}
}

func TestOpenAPITransitionalContractBudget(t *testing.T) {
	const maximumTransitionalContracts = 329
	contents, err := os.ReadFile("../../../docs/openapi/openapi.yaml")
	if err != nil {
		t.Fatal(err)
	}
	count := strings.Count(string(contents), "x-clpr-router-derived: true")
	if count > maximumTransitionalContracts {
		t.Fatalf("transitional OpenAPI contracts increased from budget %d to %d", maximumTransitionalContracts, count)
	}
}

func appendGeneratedOperations(filename string, routes []string) error {
	contents, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	marker := []byte("\ncomponents:\n")
	position := strings.Index(string(contents), string(marker))
	if position < 0 {
		return fmt.Errorf("OpenAPI components marker not found")
	}

	grouped := make(map[string][]string)
	for _, route := range routes {
		parts := strings.SplitN(route, " ", 2)
		grouped[parts[1]] = append(grouped[parts[1]], parts[0])
	}
	paths := make([]string, 0, len(grouped))
	for path := range grouped {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	var generated strings.Builder
	for _, path := range paths {
		generated.WriteString("\n  " + path + ":\n")
		for _, upperMethod := range grouped[path] {
			method := strings.ToLower(upperMethod)
			operationID := strings.Trim(operationIDPart.ReplaceAllString(upperMethod+" "+path, " "), " ")
			operationID = strings.ReplaceAll(strings.Title(strings.ToLower(operationID)), " ", "") //nolint:staticcheck -- deterministic generator
			operationID = method + strings.TrimPrefix(operationID, strings.Title(method))
			generated.WriteString("    " + method + ":\n")
			generated.WriteString("      tags: [Generated Route Contracts]\n")
			generated.WriteString("      summary: " + upperMethod + " " + path + "\n")
			generated.WriteString("      description: Router-derived operation pending a route-specific response schema.\n")
			generated.WriteString("      operationId: " + operationID + "\n")
			generated.WriteString("      x-clpr-router-derived: true\n")
			parameters := openAPIParameter.FindAllStringSubmatch(path, -1)
			if len(parameters) > 0 {
				generated.WriteString("      parameters:\n")
				for _, parameter := range parameters {
					generated.WriteString("        - name: " + parameter[1] + "\n")
					generated.WriteString("          in: path\n          required: true\n          schema:\n            type: string\n")
				}
			}
			generated.WriteString("      responses:\n")
			generated.WriteString("        '200':\n          $ref: '#/components/responses/RouterDerivedSuccess'\n")
			generated.WriteString("        '400':\n          $ref: '#/components/responses/BadRequest'\n")
			generated.WriteString("        '401':\n          $ref: '#/components/responses/Unauthorized'\n")
			generated.WriteString("        '500':\n          $ref: '#/components/responses/InternalServerError'\n")
		}
	}

	updated := append([]byte{}, contents[:position]...)
	updated = append(updated, []byte(generated.String())...)
	updated = append(updated, contents[position:]...)
	return os.WriteFile(filename, updated, 0o644)
}

func removeGeneratedOperations(filename string) error {
	contents, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	markerPosition := strings.Index(string(contents), "      x-clpr-router-derived: true")
	if markerPosition < 0 {
		return nil
	}
	start := strings.LastIndex(string(contents[:markerPosition]), "\n  /")
	end := strings.Index(string(contents[markerPosition:]), "\ncomponents:\n")
	if start < 0 || end < 0 {
		return fmt.Errorf("could not locate generated OpenAPI operation region")
	}
	end += markerPosition
	updated := append([]byte{}, contents[:start]...)
	updated = append(updated, contents[end:]...)
	return os.WriteFile(filename, updated, 0o644)
}
