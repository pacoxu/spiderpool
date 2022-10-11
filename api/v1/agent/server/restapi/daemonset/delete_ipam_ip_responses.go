// Code generated by go-swagger; DO NOT EDIT.

// Copyright 2022 Authors of spidernet-io
// SPDX-License-Identifier: Apache-2.0

package daemonset

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/spidernet-io/spiderpool/api/v1/agent/models"
)

// DeleteIpamIPOKCode is the HTTP code returned for type DeleteIpamIPOK
const DeleteIpamIPOKCode int = 200

/*
DeleteIpamIPOK Success

swagger:response deleteIpamIpOK
*/
type DeleteIpamIPOK struct {
}

// NewDeleteIpamIPOK creates DeleteIpamIPOK with default headers values
func NewDeleteIpamIPOK() *DeleteIpamIPOK {

	return &DeleteIpamIPOK{}
}

// WriteResponse to the client
func (o *DeleteIpamIPOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.Header().Del(runtime.HeaderContentType) //Remove Content-Type on empty responses

	rw.WriteHeader(200)
}

// DeleteIpamIPFailureCode is the HTTP code returned for type DeleteIpamIPFailure
const DeleteIpamIPFailureCode int = 500

/*
DeleteIpamIPFailure Addresses release failure

swagger:response deleteIpamIpFailure
*/
type DeleteIpamIPFailure struct {

	/*
	  In: Body
	*/
	Payload models.Error `json:"body,omitempty"`
}

// NewDeleteIpamIPFailure creates DeleteIpamIPFailure with default headers values
func NewDeleteIpamIPFailure() *DeleteIpamIPFailure {

	return &DeleteIpamIPFailure{}
}

// WithPayload adds the payload to the delete ipam Ip failure response
func (o *DeleteIpamIPFailure) WithPayload(payload models.Error) *DeleteIpamIPFailure {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the delete ipam Ip failure response
func (o *DeleteIpamIPFailure) SetPayload(payload models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *DeleteIpamIPFailure) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(500)
	payload := o.Payload
	if err := producer.Produce(rw, payload); err != nil {
		panic(err) // let the recovery middleware deal with this
	}
}
