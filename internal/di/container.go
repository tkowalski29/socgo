package di

import (
	"fmt"
	"net/http"
	"reflect"
)

type Container struct {
	services map[string]interface{}
}

func NewContainer() *Container {
	return &Container{
		services: make(map[string]interface{}),
	}
}

func (c *Container) Register(name string, service interface{}) {
	c.services[name] = service
}

func (c *Container) Get(name string) (interface{}, error) {
	service, exists := c.services[name]
	if !exists {
		return nil, fmt.Errorf("service %s not found", name)
	}
	return service, nil
}

func (c *Container) GetHTTPHandler(name string) (http.Handler, error) {
	service, err := c.Get(name)
	if err != nil {
		return nil, err
	}
	
	handler, ok := service.(http.Handler)
	if !ok {
		return nil, fmt.Errorf("service %s is not an http.Handler", name)
	}
	
	return handler, nil
}

func (c *Container) MustGet(name string) interface{} {
	service, err := c.Get(name)
	if err != nil {
		panic(err)
	}
	return service
}

func (c *Container) Has(name string) bool {
	_, exists := c.services[name]
	return exists
}

func (c *Container) GetType(name string) (reflect.Type, error) {
	service, err := c.Get(name)
	if err != nil {
		return nil, err
	}
	return reflect.TypeOf(service), nil
}

func (c *Container) List() []string {
	keys := make([]string, 0, len(c.services))
	for key := range c.services {
		keys = append(keys, key)
	}
	return keys
}