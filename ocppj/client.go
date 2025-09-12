package ocppj

import (
	"context"
	"errors"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"gopkg.in/go-playground/validator.v9"

	"github.com/lorenzodonini/ocpp-go/ocpp"
	"github.com/lorenzodonini/ocpp-go/ws"
)

// The endpoint initiating the connection to an OCPP server, in an OCPP-J topology.
// During message exchange, the two roles may be reversed (depending on the message direction), but a client struct remains associated to a charge point/charging station.
type Client struct {
	Endpoint
	client                ws.Client
	Id                    string
	requestHandler        func(ctx context.Context, request ocpp.Request, requestId string, action string)
	responseHandler       func(ctx context.Context, response ocpp.Response, requestId string)
	errorHandler          func(ctx context.Context, err *ocpp.Error, details interface{})
	onDisconnectedHandler func(err error)
	onReconnectedHandler  func()
	invalidMessageHook    func(err *ocpp.Error, rawMessage string, parsedFields []interface{}) *ocpp.Error
	dispatcher            ClientDispatcher
	RequestState          ClientState
	contextMap            *ContextMap
}

// Creates a new Client endpoint.
// Requires a unique client ID, a websocket client, a struct for queueing/dispatching requests,
// a state handler and a list of supported profiles (optional).
//
// You may create a simple new server by using these default values:
//
//	s := ocppj.NewClient(ws.NewClient(), nil, nil)
//
// The wsClient parameter cannot be nil. Refer to the ws package for information on how to create and
// customize a websocket client.
func NewClient(id string, wsClient ws.Client, dispatcher ClientDispatcher, stateHandler ClientState, profiles ...*ocpp.Profile) *Client {
	endpoint := Endpoint{}
	if wsClient == nil {
		panic("wsClient parameter cannot be nil")
	}
	for _, profile := range profiles {
		endpoint.AddProfile(profile)
	}
	if dispatcher == nil {
		dispatcher = NewDefaultClientDispatcher(NewFIFOClientQueue(10))
	}
	if stateHandler == nil {
		stateHandler = NewClientState()
	}
	dispatcher.SetNetworkClient(wsClient)
	dispatcher.SetPendingRequestState(stateHandler)

	client := &Client{Endpoint: endpoint, client: wsClient, Id: id, dispatcher: dispatcher, RequestState: stateHandler, contextMap: NewContextMap()}

	// Set up default timeout handler to clean up context map and close spans
	client.SetOnRequestCanceled(nil) // This will set up the cleanup handler

	return client
}

// Return incoming requests handler.
func (c *Client) GetRequestHandler() func(ctx context.Context, request ocpp.Request, requestId string, action string) {
	return c.requestHandler
}

// Registers a handler for incoming requests.
func (c *Client) SetRequestHandler(handler func(ctx context.Context, request ocpp.Request, requestId string, action string)) {
	c.requestHandler = handler
}

// Return incoming responses handler.
func (c *Client) GetResponseHandler() func(ctx context.Context, response ocpp.Response, requestId string) {
	return c.responseHandler
}

// Registers a handler for incoming responses.
func (c *Client) SetResponseHandler(handler func(ctx context.Context, response ocpp.Response, requestId string)) {
	c.responseHandler = handler
}

// Return incoming error messages handler.
func (c *Client) GetErrorHandler() func(ctx context.Context, err *ocpp.Error, details interface{}) {
	return c.errorHandler
}

// Registers a handler for incoming error messages.
func (c *Client) SetErrorHandler(handler func(ctx context.Context, err *ocpp.Error, details interface{})) {
	c.errorHandler = handler
}

// SetInvalidMessageHook registers an optional hook for incoming messages that couldn't be parsed.
// This hook is called when a message is received but cannot be parsed to the target OCPP message struct.
//
// The application is notified synchronously of the error.
// The callback provides the raw JSON string, along with the parsed fields.
// The application MUST return as soon as possible, since the hook is called synchronously and awaits a return value.
//
// While the hook does not allow responding to the message directly,
// the return value will be used to send an OCPP error to the other endpoint.
//
// If no handler is registered (or no error is returned by the hook),
// the internal error message is sent to the client without further processing.
//
// Note: Failing to return from the hook will cause the client to block indefinitely.
func (c *Client) SetInvalidMessageHook(hook func(err *ocpp.Error, rawMessage string, parsedFields []interface{}) *ocpp.Error) {
	c.invalidMessageHook = hook
}

func (c *Client) SetOnDisconnectedHandler(handler func(err error)) {
	c.onDisconnectedHandler = handler
}

func (c *Client) SetOnReconnectedHandler(handler func()) {
	c.onReconnectedHandler = handler
}

// Registers the handler to be called on timeout.
func (c *Client) SetOnRequestCanceled(handler func(requestId string, request ocpp.Request, err *ocpp.Error)) {
	// Chain the user handler with our internal cleanup handler
	c.dispatcher.SetOnRequestCanceled(func(requestId string, request ocpp.Request, err *ocpp.Error) {
		// First, clean up our internal context map
		if _, storedSpan, exists := c.contextMap.LoadAndDelete(requestId); exists {
			storedSpan.RecordError(fmt.Errorf("request timed out: %s", err.Description))
			storedSpan.End()
			log.Debugf("cleaned up context and span for timed out request %s", requestId)
		}
		// Then call the user's handler if provided
		if handler != nil {
			handler(requestId, request, err)
		}
	})
}

// Connects to the given serverURL and starts running the I/O loop for the underlying connection.
//
// If the connection is established successfully, the function returns control to the caller immediately.
// The read/write routines are run on dedicated goroutines, so the main thread can perform other operations.
//
// In case of disconnection, the client handles re-connection automatically.
// The client will attempt to re-connect to the server forever, until it is stopped by invoking the Stop method.
//
// An error may be returned, if establishing the connection failed.
func (c *Client) Start(serverURL string) error {
	// Set internal message handler
	c.client.SetMessageHandler(c.ocppMessageHandler)
	c.client.SetDisconnectedHandler(c.onDisconnected)
	c.client.SetReconnectedHandler(c.onReconnected)
	// Connect & run
	fullUrl := fmt.Sprintf("%v/%v", serverURL, c.Id)
	err := c.client.Start(fullUrl)
	if err == nil {
		c.dispatcher.Start()
	}
	return err
}

func (c *Client) StartWithRetries(serverURL string) {
	// Set internal message handler
	c.client.SetMessageHandler(c.ocppMessageHandler)
	c.client.SetDisconnectedHandler(c.onDisconnected)
	c.client.SetReconnectedHandler(c.onReconnected)
	// Connect & run
	fullUrl := fmt.Sprintf("%v/%v", serverURL, c.Id)
	c.client.StartWithRetries(fullUrl)
	c.dispatcher.Start()
}

// Stops the client.
// The underlying I/O loop is stopped and all pending requests are cleared.
func (c *Client) Stop() {
	// Overwrite handler to intercept disconnected signal
	cleanupC := make(chan struct{}, 1)
	if c.IsConnected() {
		c.client.SetDisconnectedHandler(func(err error) {
			cleanupC <- struct{}{}
		})
	} else {
		close(cleanupC)
	}
	c.client.Stop()
	if c.dispatcher.IsRunning() {
		c.dispatcher.Stop()
	}
	// Wait for websocket to be cleaned up
	<-cleanupC
}

func (c *Client) IsConnected() bool {
	return c.client.IsConnected()
}

// Sends an OCPP Request to the server.
// The protocol is based on request-response and cannot send multiple messages concurrently.
// To guarantee this, outgoing messages are added to a queue and processed sequentially.
//
// Returns an error in the following cases:
//
// - the client wasn't started
//
// - message validation fails (request is malformed)
//
// - the endpoint doesn't support the feature
//
// - the output queue is full
func (c *Client) SendRequest(request ocpp.Request) error {
	return c.SendRequestWithContext(context.Background(), request)
}

func (c *Client) SendRequestWithContext(ctx context.Context, request ocpp.Request) error {
	tracer := otel.Tracer("ocpp-go/ocppj")
	ctx, span := tracer.Start(ctx, fmt.Sprintf("cs.out_call.%s", request.GetFeatureName()))

	if !c.dispatcher.IsRunning() {
		err := fmt.Errorf("ocppj client is not started, couldn't send request")
		span.RecordError(err)
		span.End()
		return err
	}
	call, err := c.CreateCall(request)
	if err != nil {
		span.RecordError(err)
		span.End()
		return err
	}

	span.SetAttributes(
		attribute.String("ocpp.unique_id", call.UniqueId),
		attribute.String("ocpp.msg_type", "call"),
		attribute.String("ocpp.action", call.Action),
		attribute.String("ocpp.role", "charging_station"),
	)

	jsonMessage, err := call.MarshalJSON()
	if err != nil {
		span.RecordError(err)
		span.End()
		return err
	}
	// Store context and span for later retrieval when response is received
	c.contextMap.Store(call.UniqueId, ctx, span)

	// Message will be processed by dispatcher. A dedicated mechanism allows to delegate the message queue handling.
	if err = c.dispatcher.SendRequestWithContext(ctx, RequestBundle{Call: call, Data: jsonMessage, Ctx: ctx}); err != nil {
		// Remove context from map if dispatch fails
		c.contextMap.Delete(call.UniqueId)
		log.Errorf("error dispatching request [%s, %s]: %v", call.UniqueId, call.Action, err)
		span.RecordError(err)
		span.End()
		return err
	}
	log.Debugf("enqueued CALL [%s, %s]", call.UniqueId, call.Action)
	return nil
}

// Sends an OCPP Response to the server.
// The requestID parameter is required and identifies the previously received request.
//
// Returns an error in the following cases:
//
// - message validation fails (response is malformed)
//
// - the endpoint doesn't support the feature
//
// - a network error occurred
func (c *Client) SendResponse(requestId string, response ocpp.Response) error {
	return c.SendResponseWithContext(context.Background(), requestId, response)
}

func (c *Client) SendResponseWithContext(ctx context.Context, requestId string, response ocpp.Response) error {
	tracer := otel.Tracer("ocppj.client")
	ctx, span := tracer.Start(ctx, fmt.Sprintf("cs.out_call_result.%s", response.GetFeatureName()))
	defer span.End()

	span.SetAttributes(
		attribute.String("ocpp.unique_id", requestId),
		attribute.String("ocpp.msg_type", "call_result"),
		attribute.String("ocpp.action", response.GetFeatureName()),
		attribute.String("ocpp.role", "charging_station"),
	)

	callResult, err := c.CreateCallResult(response, requestId)
	if err != nil {
		span.RecordError(err)
		return err
	}
	jsonMessage, err := callResult.MarshalJSON()
	if err != nil {
		err = ocpp.NewError(GenericError, err.Error(), requestId)
		span.RecordError(err)
		return err
	}
	if err = c.client.WriteWithContext(ctx, jsonMessage); err != nil {
		log.Errorf("error sending response [%s]: %v", callResult.GetUniqueId(), err)
		err = ocpp.NewError(GenericError, err.Error(), requestId)
		span.RecordError(err)
		return err
	}
	log.Debugf("sent CALL RESULT [%s]", callResult.GetUniqueId())
	log.Debugf("sent JSON message to server: %s", string(jsonMessage))
	return nil
}

// Sends an OCPP Error to the server.
// The requestID parameter is required and identifies the previously received request.
//
// Returns an error in the following cases:
//
// - message validation fails (error is malformed)
//
// - a network error occurred
func (c *Client) SendError(requestId string, errorCode ocpp.ErrorCode, description string, details interface{}) error {
	return c.SendErrorWithContext(context.Background(), requestId, errorCode, description, details)
}

func (c *Client) SendErrorWithContext(ctx context.Context, requestId string, errorCode ocpp.ErrorCode, description string, details interface{}) error {
	tracer := otel.Tracer("ocppj.client")
	ctx, span := tracer.Start(ctx, "cs.out_call_error")
	defer span.End()

	span.SetAttributes(
		attribute.String("ocpp.unique_id", requestId),
		attribute.String("ocpp.msg_type", "call_error"),
		attribute.String("ocpp.role", "charging_station"),
		attribute.String("ocpp.error.code", string(errorCode)),
		attribute.String("ocpp.error.description", description),
	)

	callError, err := c.CreateCallError(requestId, errorCode, description, details)
	if err != nil {
		span.RecordError(err)
		return err
	}
	jsonMessage, err := callError.MarshalJSON()
	if err != nil {
		err = ocpp.NewError(GenericError, err.Error(), requestId)
		span.RecordError(err)
		return err
	}
	if err = c.client.WriteWithContext(ctx, jsonMessage); err != nil {
		log.Errorf("error sending response error [%s]: %v", callError.UniqueId, err)
		err = ocpp.NewError(GenericError, err.Error(), requestId)
		span.RecordError(err)
		return err
	}
	log.Debugf("sent CALL ERROR [%s]", callError.UniqueId)
	log.Debugf("sent JSON message to server: %s", string(jsonMessage))
	return nil
}

func (c *Client) ocppMessageHandler(ctx context.Context, data []byte) error {
	parsedJson, err := ParseRawJsonMessage(data)
	if err != nil {
		log.Error(err)
		return err
	}
	log.Debugf("received JSON message from server: %s", string(data))
	message, err := c.ParseMessage(parsedJson, c.RequestState)
	if err != nil {
		ocppErr := err.(*ocpp.Error)
		messageID := ocppErr.MessageId
		// Support ad-hoc callback for invalid message handling
		if c.invalidMessageHook != nil {
			err2 := c.invalidMessageHook(ocppErr, string(data), parsedJson)
			// If the hook returns an error, use it as output error. If not, use the original error.
			if err2 != nil {
				ocppErr = err2
				ocppErr.MessageId = messageID
			}
		}
		err = ocppErr
		// Send error to other endpoint if a message ID is available
		if ocppErr.MessageId != "" {
			err2 := c.SendError(ocppErr.MessageId, ocppErr.Code, ocppErr.Description, nil)
			if err2 != nil {
				return err2
			}
		}
		log.Error(err)
		return err
	}
	if message != nil {
		switch message.GetMessageTypeId() {
		case CALL:
			call := message.(*Call)
			log.Debugf("handling incoming CALL [%s, %s]", call.UniqueId, call.Action)
			c.requestHandler(ctx, call.Payload, call.UniqueId, call.Action)
		case CALL_RESULT:
			callResult := message.(*CallResult)
			log.Debugf("handling incoming CALL RESULT [%s]", callResult.UniqueId)

			// Retrieve and clean up stored context and span
			storedCtx, storedSpan, exists := c.contextMap.LoadAndDelete(callResult.UniqueId)
			if exists {
				// Close the original request span
				storedSpan.End()

				// Create a child span for response handling using the stored context
				tracer := otel.Tracer("ocppj.client")
				_, span := tracer.Start(storedCtx, "cs.handle_call_result")
				defer span.End()
				span.SetAttributes(
					attribute.String("response.id", callResult.UniqueId),
					attribute.String("client.id", c.Id),
				)
			}

			c.dispatcher.CompleteRequest(callResult.GetUniqueId()) // Remove current request from queue and send next one
			if c.responseHandler != nil {
				c.responseHandler(ctx, callResult.Payload, callResult.UniqueId)
			}
		case CALL_ERROR:
			callError := message.(*CallError)
			log.Debugf("handling incoming CALL ERROR [%s]", callError.UniqueId)

			// Retrieve and clean up stored context and span
			if storedCtx, storedSpan, exists := c.contextMap.LoadAndDelete(callError.UniqueId); exists {
				// Close the original request span with error status
				storedSpan.RecordError(fmt.Errorf("OCPP call error: %s - %s", callError.ErrorCode, callError.ErrorDescription))
				storedSpan.End()

				// Create a child span for error handling using the stored context
				tracer := otel.Tracer("ocppj.client")
				_, span := tracer.Start(storedCtx, "cs.handle_call_error")
				span.SetAttributes(
					attribute.String("error.id", callError.UniqueId),
					attribute.String("error.code", string(callError.ErrorCode)),
					attribute.String("error.description", callError.ErrorDescription),
					attribute.String("client.id", c.Id),
				)
				span.End()
			}

			c.dispatcher.CompleteRequest(callError.GetUniqueId()) // Remove current request from queue and send next one
			if c.errorHandler != nil {
				c.errorHandler(ctx, ocpp.NewError(callError.ErrorCode, callError.ErrorDescription, callError.UniqueId), callError.ErrorDetails)
			}
		}
	}
	return nil
}

// HandleFailedResponseError allows to handle failures while sending responses (either CALL_RESULT or CALL_ERROR).
// It internally analyzes and creates an ocpp.Error based on the given error.
// It will the attempt to send it to the server.
//
// The function helps to prevent starvation on the other endpoint, which is caused by a response never reaching it.
// The method will, however, only attempt to send a default error once.
// If this operation fails, the other endpoint may still starve.
func (c *Client) HandleFailedResponseError(requestID string, err error, featureName string) {
	log.Debugf("handling error for failed response [%s]", requestID)
	var responseErr *ocpp.Error
	// There's several possible errors: invalid profile, invalid payload or send error
	switch err.(type) {
	case validator.ValidationErrors:
		// Validation error
		var validationErr validator.ValidationErrors
		errors.As(err, &validationErr)
		responseErr = errorFromValidation(c, validationErr, requestID, featureName)
	case *ocpp.Error:
		// Internal OCPP error
		errors.As(err, &responseErr)
	case error:
		// Unknown error
		responseErr = ocpp.NewError(GenericError, err.Error(), requestID)
	}
	// Send an OCPP error to the target, since no regular response could be sent
	_ = c.SendError(requestID, responseErr.Code, responseErr.Description, nil)
}

func (c *Client) onDisconnected(err error) {
	log.Error("disconnected from server", err)
	c.dispatcher.Pause()
	if c.onDisconnectedHandler != nil {
		c.onDisconnectedHandler(err)
	}
}

func (c *Client) onReconnected() {
	if c.onReconnectedHandler != nil {
		c.onReconnectedHandler()
	}
	c.dispatcher.Resume()
}
