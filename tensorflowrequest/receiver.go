/*
Copyright (c) 2021 TriggerMesh Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
)

const (
	tensformationEvent = "io.triggermesh.transformations.tensformation.response"
	response           = "io.triggermesh.transformations.tensorflowrequest.response"
)

func (recv *Receiver) receive(ctx context.Context, e cloudevents.Event) (*cloudevents.Event, cloudevents.Result) {
	log.Printf("Processing event from source %q", e.Source())
	if typ := e.Type(); typ != tensformationEvent {
		fmt.Println("wrong event type")
		return emitErrorEvent("wrong event type", "wrongEventType")
	}

	req := &B64ResponseEvent{}
	if err := e.DataAs(&req); err != nil {
		log.Print(err)
		return emitErrorEvent(err.Error(), "unmarshalingEvent")
	}

	err, tfResponse := recv.makeTensorflowRequest(req.B64)
	if err != nil {
		log.Print(err)
		return emitErrorEvent(err.Error(), "requestingFromTensorflow")
	}

	event := cloudevents.NewEvent(cloudevents.VersionV1)
	event.SetType(response)
	event.SetSource(req.URL)
	event.SetTime(time.Now())
	err = event.SetData(cloudevents.ApplicationJSON, tfResponse)
	if err != nil {
		log.Print(err)
		return emitErrorEvent(err.Error(), "settingCEData")
	}

	return &event, cloudevents.ResultACK
}

func (recv *Receiver) makeTensorflowRequest(image string) (error, []byte) {
	reqBody := &TensorflowRequest{
		Instances: []struct {
			B64 string "json:\"b64\""
		}{{B64: image}},
	}

	b, err := json.Marshal(reqBody)
	if err != nil {
		return err, b
	}

	request, err := http.NewRequest(http.MethodPost, recv.tfEndpoint, bytes.NewBuffer(b))
	if err != nil {
		return err, b
	}

	res, err := recv.httpClient.Do(request)
	if err != nil {
		return err, b
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err, b
	}

	return nil, body
}

func emitErrorEvent(er string, source string) (*cloudevents.Event, cloudevents.Result) {
	responseEvent := cloudevents.NewEvent(cloudevents.VersionV1)
	responseEvent.SetType(response + ".error")
	responseEvent.SetSource(source)
	responseEvent.SetTime(time.Now())
	err := responseEvent.SetData(cloudevents.ApplicationJSON, er)
	if err != nil {
		log.Print(err)
		return nil, cloudevents.NewHTTPResult(http.StatusInternalServerError, "setting cloudevent response data")
	}

	return &responseEvent, cloudevents.NewHTTPResult(http.StatusInternalServerError, er)
}
