package main

import (
	"context"
	"fmt"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/legacy"
	"log"
	"net/http"
)

func main() {
	log.Println("Hello world :D")

	loader := openapi3.NewLoader()
	spec, err := loader.LoadFromFile("./lib/example.yml")
	if err != nil {
		panic(err)
	}

	err = spec.Validate(context.Background())
	if err != nil {
		panic(err)
	}

	router, _ := legacy.NewRouter(spec)


	httpClient := http.Client{}

	for _, server := range spec.Servers {
		baseUrl := server.URL

		for pathName, _ := range spec.Paths {
			err := testPath(baseUrl, pathName, router, httpClient)
			if err != nil {
				panic(err)
			}
		}
	}
}

func testPath(baseUrl string, endpoint string, router routers.Router, httpClient http.Client) error {
	url := baseUrl + endpoint + "?id=123"
	log.Printf("%q: about to do GET to %q", endpoint, url)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	res, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()


	log.Printf("status code was %d", res.StatusCode)
	route, pathParams, err := router.FindRoute(req)
	if err != nil {
		return err
	}

	if route.Operation.Responses.Get(res.StatusCode) == nil {
		panic(fmt.Errorf("no response for status code %d", res.StatusCode))
	}


	responseValidationInput := &openapi3filter.ResponseValidationInput{
		RequestValidationInput: &openapi3filter.RequestValidationInput{
			Request: req,
			Route: route,
			PathParams: pathParams,
		},
		Status:                 res.StatusCode,
		Header:                 res.Header,
		Body:                   res.Body,
		Options: &openapi3filter.Options{IncludeResponseStatus: true},
		//Options: &openapi3filter.Options{
		//	ExcludeRequestBody:    true,
		//	ExcludeResponseBody:   false,
		//	IncludeResponseStatus: false,
		//	MultiError:            false,
		//	AuthenticationFunc:    nil,
		//},
	}

	ctx := context.Background()
	if err := openapi3filter.ValidateResponse(ctx, responseValidationInput); err != nil {
		return err
	}

	log.Println("YAY :D EVERYTHING WORKED!!!")

	//matchedResponse := path.Get.Responses[fmt.Sprintf("%d", res.StatusCode)]
	//log.Printf("%+v", matchedResponse)
	//
	//panic("yay :D")
	return nil
}
