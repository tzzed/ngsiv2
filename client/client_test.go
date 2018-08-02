package client_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/phoops/ngsiv2/client"
	"github.com/phoops/ngsiv2/model"
)

var apiResourcesHandler = func(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, `{"entities_url":"/v2/entities","types_url":"/v2/types","subscriptions_url":"/v2/subscriptions","registrations_url":"/v2/registrations"}`)
}

func TestBatchUpdateBadRequest(t *testing.T) {
	ts := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprintf(w, `{"error":"ParseError","description":"Errors found in incoming JSON buffer"}`)
			}))
	defer ts.Close()

	cli, err := client.NewNgsiV2Client(client.SetUrl(ts.URL))
	if err != nil {
		t.Fatalf("Unexpected error: '%v'", err)
	}
	if err := cli.BatchUpdate(&model.BatchUpdate{}); err == nil {
		t.Fatal("Expected an error")
	}
}

func TestBatchUpdateNoContent(t *testing.T) {
	ts := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("Accept") != "application/json" {
					t.Fatal("Missing application/json accept header")
				}
				if b, err := ioutil.ReadAll(r.Body); err != nil {
					t.Fatalf("Unexpected error: '%v'", err)
				} else if !strings.Contains(string(b), "entities") {
					t.Fatal("Request doesn't contain entities")
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusNoContent)
			}))
	defer ts.Close()

	cli, err := client.NewNgsiV2Client(client.SetUrl(ts.URL))
	if err != nil {
		t.Fatalf("Unexpected error: '%v'", err)
	}
	if err := cli.BatchUpdate(&model.BatchUpdate{}); err != nil {
		t.Fatalf("Unexpected error: '%v'", err)
	}
}

func TestRetrieveAPIResources(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(apiResourcesHandler))
	defer ts.Close()

	cli, err := client.NewNgsiV2Client(client.SetUrl(ts.URL))
	if err != nil {
		t.Fatalf("Unexpected error: '%v'", err)
	}

	if res, err := cli.RetrieveAPIResources(); err != nil {
		t.Fatalf("Unexpected error: '%v'", err)
	} else {
		if res.EntitiesUrl != "/v2/entities" ||
			res.TypesUrl != "/v2/types" ||
			res.SubscriptionsUrl != "/v2/subscriptions" ||
			res.RegistrationsUrl != "/v2/registrations" {
			t.Fatal("Failed reading API resources values")
		}
	}
}

func TestRetrieveEntity(t *testing.T) {
	ts := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				if strings.HasSuffix(r.URL.Path, "/v2") {
					apiResourcesHandler(w, r)
				} else {
					if r.Header.Get("Accept") != "application/json" {
						t.Fatal("Missing application/json accept header")
					}
					if r.Header.Get("Content-Type") != "" {
						t.Fatal("No Content-Type allowed for GET request")
					}
					if !strings.HasSuffix(r.URL.Path, "/r1") {
						t.Fatal("Expected 'r1' as id")
					}
					if r.URL.Query().Get("type") != "Room" {
						t.Fatalf("Expected 'type' value: 'Room', got '%s'", r.URL.Query().Get("type"))
					}
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					fmt.Fprint(w, `{"id":"r1","type":"Room","pressure":{"type":"Integer","value":"720","metadata":{}},"temperature":{"type":"Float","value":23,"metadata":{}}}`)
				}
			}))
	defer ts.Close()

	cli, err := client.NewNgsiV2Client(client.SetUrl(ts.URL))
	if err != nil {
		t.Fatalf("Unexpected error: '%v'", err)
	}

	if res, err := cli.RetrieveEntity("r1", client.RetrieveEntitySetType("Room")); err != nil {
		t.Fatalf("Unexpected error: '%v'", err)
	} else {
		if res.Id != "r1" ||
			res.Type != "Room" ||
			res.Attributes["temperature"].Type != model.FloatType {
			t.Fatal("Invalid entity retrieved")
		}
	}
}

func TestRetrieveEntities(t *testing.T) {
	ts := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				if strings.HasSuffix(r.URL.Path, "/v2") {
					apiResourcesHandler(w, r)
				} else {
					if r.Header.Get("Accept") != "application/json" {
						t.Fatal("Missing application/json accept header")
					}
					if r.Header.Get("Content-Type") != "" {
						t.Fatal("No Content-Type allowed for GET request")
					}
					if r.URL.Query().Get("type") != "Room" {
						t.Fatalf("Expected 'type' value: 'Room', got '%s'", r.URL.Query().Get("type"))
					}
					if r.URL.Query().Get("q") != "temperature>30" {
						t.Fatalf("Expected 'q' expression: 'temperature>30', got '%s'", r.URL.Query().Get("q"))
					}
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					fmt.Fprint(w, `[{"id":"r2","type":"Room","pressure":{"type":"Integer","value":"720","metadata":{}},"temperature":{"type":"Float","value":34,"metadata":{}}},{"id":"r5","type":"Room","pressure":{"type":"Integer","value":"700","metadata":{}},"temperature":{"type":"Float","value":31,"metadata":{}}}
]`)
				}
			}))
	defer ts.Close()

	cli, err := client.NewNgsiV2Client(client.SetUrl(ts.URL))
	if err != nil {
		t.Fatalf("Unexpected error: '%v'", err)
	}

	qst, err := model.NewBinarySimpleQueryStatement("temperature", model.SQGreaterThan, "30")
	if err != nil {
		t.Fatalf("Unexpected error: '%v'", err)
	}
	if res, err := cli.ListEntities(
		client.ListEntitiesSetType("Room"),
		client.ListEntitiesAddQueryStatement(qst)); err != nil {
		t.Fatalf("Unexpected error: '%v'", err)
	} else {
		if len(res) != 2 {
			t.Fatalf("Expected 2 entities, got %d", len(res))
		}
		if res[0].Id != "r2" ||
			res[0].Type != "Room" ||
			res[0].Attributes["temperature"].Type != model.FloatType ||
			res[1].Id != "r5" ||
			res[1].Type != "Room" ||
			res[1].Attributes["temperature"].Type != model.FloatType {
			t.Fatal("Invalid entities retrieved")
		}
	}
}
