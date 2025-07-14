package di

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestContainer_RegisterAndGet(t *testing.T) {
	container := NewContainer()
	testService := "test_service"
	
	container.Register("test", testService)
	
	service, err := container.Get("test")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	
	if service != testService {
		t.Errorf("Expected %v, got %v", testService, service)
	}
}

func TestContainer_GetNonExistent(t *testing.T) {
	container := NewContainer()
	
	_, err := container.Get("non_existent")
	if err == nil {
		t.Error("Expected error for non-existent service")
	}
}

func TestContainer_GetHTTPHandler(t *testing.T) {
	container := NewContainer()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	
	container.Register("handler", handler)
	
	retrievedHandler, err := container.GetHTTPHandler("handler")
	if err != nil {
		t.Fatalf("GetHTTPHandler() error = %v", err)
	}
	
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	retrievedHandler.ServeHTTP(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestContainer_GetHTTPHandlerWrongType(t *testing.T) {
	container := NewContainer()
	container.Register("not_handler", "string")
	
	_, err := container.GetHTTPHandler("not_handler")
	if err == nil {
		t.Error("Expected error for non-handler service")
	}
}

func TestContainer_MustGet(t *testing.T) {
	container := NewContainer()
	testService := "test_service"
	
	container.Register("test", testService)
	
	service := container.MustGet("test")
	if service != testService {
		t.Errorf("Expected %v, got %v", testService, service)
	}
}

func TestContainer_MustGetPanic(t *testing.T) {
	container := NewContainer()
	
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for non-existent service")
		}
	}()
	
	container.MustGet("non_existent")
}

func TestContainer_Has(t *testing.T) {
	container := NewContainer()
	container.Register("test", "service")
	
	if !container.Has("test") {
		t.Error("Expected Has() to return true for existing service")
	}
	
	if container.Has("non_existent") {
		t.Error("Expected Has() to return false for non-existent service")
	}
}

func TestContainer_List(t *testing.T) {
	container := NewContainer()
	container.Register("service1", "value1")
	container.Register("service2", "value2")
	
	services := container.List()
	if len(services) != 2 {
		t.Errorf("Expected 2 services, got %d", len(services))
	}
	
	hasService1 := false
	hasService2 := false
	for _, service := range services {
		if service == "service1" {
			hasService1 = true
		}
		if service == "service2" {
			hasService2 = true
		}
	}
	
	if !hasService1 || !hasService2 {
		t.Error("Expected both services to be listed")
	}
}