package gql

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"bitbucket.org/zwzn/comicbox/comicboxd/app"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
)

var GQLHandler *handler.Handler

type CloserBuffer struct {
	*bytes.Buffer
}

func (cb *CloserBuffer) Close() error {
	return nil
}

func Query(r *http.Request, query string, vars map[string]interface{}, response interface{}) error {
	request := map[string]interface{}{
		// "operationName": "getBook",
		"query":     query,
		"variables": vars,
	}

	b, err := json.Marshal(request)
	if err != nil {
		return err
	}

	resp := httptest.NewRecorder()
	newR := httptest.NewRequest("POST", "http://0.0.0.0/graphql", &CloserBuffer{bytes.NewBuffer(b)})

	app.CtxSet(newR, app.Ctx(r))

	GQLHandler.ServeHTTP(resp, newR)

	qr := graphql.Result{}

	json.NewDecoder(resp.Body).Decode(&qr)
	// resp.Body.Close()

	for _, err := range qr.Errors {
		return err
	}

	data, ok := qr.Data.(map[string]interface{})
	if !ok {
		return fmt.Errorf("data not a map")
	}

	if len(data) > 1 {
		return fmt.Errorf("more than one return value")
	}
	for _, fieldData := range data {
		jData, err := json.Marshal(fieldData)
		if err != nil {
			return err
		}
		err = json.Unmarshal(jData, response)
		if err != nil {
			return err
		}
	}
	// err = mapstructure.Decode(qr.Data, response)
	// if err != nil {
	// 	return err
	// }

	return nil
}
