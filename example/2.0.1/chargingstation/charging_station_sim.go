package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"os"
	"strconv"
	"time"

	"github.com/lorenzodonini/ocpp-go/ocpp2.0.1"
	"github.com/lorenzodonini/ocpp-go/ocpp2.0.1/availability"
	"github.com/lorenzodonini/ocpp-go/ocpp2.0.1/localauth"
	"github.com/lorenzodonini/ocpp-go/ocpp2.0.1/provisioning"
	"github.com/lorenzodonini/ocpp-go/ocpp2.0.1/reservation"
	"github.com/lorenzodonini/ocpp-go/ocpp2.0.1/transactions"
	"github.com/lorenzodonini/ocpp-go/ocpp2.0.1/types"
	"github.com/lorenzodonini/ocpp-go/ocppj"
	"github.com/lorenzodonini/ocpp-go/ws"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	envVarClientID             = "CLIENT_ID"
	envVarCSMSUrl              = "CSMS_URL"
	envVarTls                  = "TLS_ENABLED"
	envVarCACertificate        = "CA_CERTIFICATE_PATH"
	envVarClientCertificate    = "CLIENT_CERTIFICATE_PATH"
	envVarClientCertificateKey = "CLIENT_CERTIFICATE_KEY_PATH"
)

type Config struct {
	OtelExporterEndpoint  string
	ServiceName           string
	DeploymentEnvironment string
	EnableTracing         bool
}

var log *logrus.Logger

// LoadOTelConfigFromEnv loads configuration from environment variables.
func LoadOTelConfigFromEnv() Config {
	cfg := Config{
		OtelExporterEndpoint:  "0.0.0.0:4317", // Default for local SigNoz
		ServiceName:           "ocpp-go/example",
		DeploymentEnvironment: "development",
		EnableTracing:         false,
	}

	if val, ok := os.LookupEnv("OTEL_EXPORTER_OTLP_ENDPOINT"); ok {
		cfg.OtelExporterEndpoint = val
	}
	if val, ok := os.LookupEnv("OTEL_SERVICE_NAME"); ok {
		cfg.ServiceName = val
	}
	if val, ok := os.LookupEnv("OTEL_DEPLOYMENT_ENVIRONMENT"); ok {
		cfg.DeploymentEnvironment = val
	}
	if val, ok := os.LookupEnv("ENABLE_TRACING"); ok {
		if pVal, err := strconv.ParseBool(val); err == nil {
			cfg.EnableTracing = pVal
		}
	}

	log.Infof("config: %+v", cfg)

	return cfg
}

func setupTracing(ctx context.Context, isEnabled bool, cfg *Config) func(ctx context.Context) error {
	shutdownFn := func(ctx context.Context) error { return nil }
	if cfg == nil || !isEnabled {
		return shutdownFn
	}

	// Setup tracing
	conn, err := grpc.NewClient(cfg.OtelExporterEndpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Errorf("failed to create gRPC connection to OTLP exporter: %v", err)
		return shutdownFn
	}
	log.Infof("created gRPC connection to OTLP exporter at %s", cfg.OtelExporterEndpoint)

	traceExporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		log.Errorf("failed to create OTLP trace exporter: %v", err)
		return shutdownFn
	}
	bsp := sdktrace.NewBatchSpanProcessor(traceExporter)
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithSpanProcessor(bsp),
		sdktrace.WithResource(resource.NewWithAttributes(
			cfg.OtelExporterEndpoint,
			semconv.ServiceName(cfg.ServiceName),
			semconv.DeploymentEnvironmentName(cfg.DeploymentEnvironment),
		)),
	)
	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	shutdownFn = tracerProvider.Shutdown
	return shutdownFn
}

func setupChargingStation(chargingStationID string) ocpp2.ChargingStation {
	return ocpp2.NewChargingStation(chargingStationID, nil, nil)
}

func setupTlsChargingStation(chargingStationID string) ocpp2.ChargingStation {
	certPool, err := x509.SystemCertPool()
	if err != nil {
		log.Fatal(err)
	}
	// Load CA cert
	caPath, ok := os.LookupEnv(envVarCACertificate)
	if ok {
		caCert, err := os.ReadFile(caPath)
		if err != nil {
			log.Warn(err)
		} else if !certPool.AppendCertsFromPEM(caCert) {
			log.Info("no ca.cert file found, will use system CA certificates")
		}
	} else {
		log.Info("no ca.cert file found, will use system CA certificates")
	}
	// Load client certificate
	clientCertPath, ok1 := os.LookupEnv(envVarClientCertificate)
	clientKeyPath, ok2 := os.LookupEnv(envVarClientCertificateKey)
	var clientCertificates []tls.Certificate
	if ok1 && ok2 {
		certificate, err := tls.LoadX509KeyPair(clientCertPath, clientKeyPath)
		if err == nil {
			clientCertificates = []tls.Certificate{certificate}
		} else {
			log.Infof("couldn't load client TLS certificate: %v", err)
		}
	}
	// Create client with TLS config
	client := ws.NewClient(ws.WithClientTLSConfig(&tls.Config{
		RootCAs:      certPool,
		Certificates: clientCertificates,
	}))
	return ocpp2.NewChargingStation(chargingStationID, nil, client)
}

// exampleRoutine simulates a charging station flow, where a dummy transaction is started.
// The simulation runs for about 5 minutes.
func exampleRoutine(ctx context.Context, chargingStation ocpp2.ChargingStation, stateHandler *ChargingStationHandler) {
	tracer := otel.Tracer("ocpp-go/ocppj")
	ctx, span := tracer.Start(ctx, "client.example_routine")
	defer span.End()
	span.AddEvent("simulation started")

	dummyClientIdToken := types.IdToken{
		IdToken: "12345",
		Type:    types.IdTokenTypeKeyCode,
	}
	// Boot
	bootResp, err := chargingStation.BootNotification(ctx, provisioning.BootReasonPowerUp, "model1", "vendor1")
	checkError(err)
	logDefault(bootResp.GetFeatureName()).Infof("status: %v, interval: %v, current time: %v", bootResp.Status, bootResp.Interval, bootResp.CurrentTime.String())
	span.AddEvent("boot notification sent")
	span.AddEvent("iterating evse")
	// Notify EVSE status
	for eID, e := range stateHandler.evse {
		updateOperationalStatus(ctx, stateHandler, eID, availability.OperationalStatusOperative)
		// Notify connector status
		for cID := range e.connectors {
			updateConnectorStatus(ctx, stateHandler, eID, cID, availability.ConnectorStatusAvailable)
		}
	}
	// Wait for some time ...
	time.Sleep(5 * time.Second)
	// Simulate charging for connector 1
	// EV is plugged in
	evseID := 1
	evse := stateHandler.evse[evseID]
	chargingConnector := 0
	updateConnectorStatus(ctx, stateHandler, evseID, chargingConnector, availability.ConnectorStatusOccupied)
	// Start transaction
	tx := transactions.Transaction{
		TransactionID: pseudoUUID(),
		ChargingState: transactions.ChargingStateEVConnected,
	}
	evseReq := types.EVSE{ID: evseID, ConnectorID: &chargingConnector}
	txEventResp, err := chargingStation.TransactionEvent(ctx, transactions.TransactionEventStarted, types.Now(), transactions.TriggerReasonCablePluggedIn, evse.nextSequence(), tx, func(request *transactions.TransactionEventRequest) {
		request.Evse = &evseReq
	})
	span.AddEvent("transaction event sent", trace.WithAttributes(attribute.String("txId", tx.TransactionID)))
	checkError(err)
	logDefault(txEventResp.GetFeatureName()).Infof("transaction %v started", tx.TransactionID)
	stateHandler.evse[evseID].currentTransaction = tx.TransactionID
	// Authorize
	authResp, err := chargingStation.Authorize(ctx, dummyClientIdToken.IdToken, types.IdTokenTypeKeyCode)
	checkError(err)
	logDefault(authResp.GetFeatureName()).Infof("status: %v %v", authResp.IdTokenInfo.Status, getExpiryDate(&authResp.IdTokenInfo))
	// Update transaction with auth info
	txEventResp, err = chargingStation.TransactionEvent(ctx, transactions.TransactionEventUpdated, types.Now(), transactions.TriggerReasonAuthorized, evse.nextSequence(), tx, func(request *transactions.TransactionEventRequest) {
		request.Evse = &evseReq
		request.IDToken = &dummyClientIdToken
	})
	checkError(err)
	logDefault(txEventResp.GetFeatureName()).Infof("transaction %v updated", tx.TransactionID)
	// Update transaction after energy offering starts
	txEventResp, err = chargingStation.TransactionEvent(ctx, transactions.TransactionEventUpdated, types.Now(), transactions.TriggerReasonChargingStateChanged, evse.nextSequence(), tx, func(request *transactions.TransactionEventRequest) {
		request.Evse = &evseReq
		request.IDToken = &dummyClientIdToken
	})
	checkError(err)
	logDefault(txEventResp.GetFeatureName()).Infof("transaction %v updated", tx.TransactionID)
	// Periodically send meter values
	var sampleInterval time.Duration = 5
	// sampleInterval, ok := stateHandler.configuration.getInt(MeterValueSampleInterval)
	// if !ok {
	//	sampleInterval = 5
	// }
	var sampledValue types.SampledValue
	for i := 0; i < 5; i++ {
		time.Sleep(time.Second * sampleInterval)
		stateHandler.meterValue += 10
		sampledValue = types.SampledValue{
			Value:     stateHandler.meterValue,
			Context:   types.ReadingContextSamplePeriodic,
			Measurand: types.MeasurandEnergyActiveExportRegister,
			Phase:     types.PhaseL3,
			Location:  types.LocationOutlet,
			UnitOfMeasure: &types.UnitOfMeasure{
				Unit: "kWh",
			},
		}
		meterValue := types.MeterValue{
			Timestamp:    types.DateTime{Time: time.Now()},
			SampledValue: []types.SampledValue{sampledValue},
		}
		// Send meter values
		txEventResp, err = chargingStation.TransactionEvent(ctx, transactions.TransactionEventUpdated, types.Now(), transactions.TriggerReasonMeterValuePeriodic, evse.nextSequence(), tx, func(request *transactions.TransactionEventRequest) {
			request.MeterValue = []types.MeterValue{meterValue}
			request.IDToken = &dummyClientIdToken
		})
		checkError(err)
		logDefault(txEventResp.GetFeatureName()).Infof("transaction %v updated with periodic meter values", tx.TransactionID)
		// Increase meter value
		stateHandler.meterValue += 2
	}
	// Stop charging for connector 1
	updateConnectorStatus(ctx, stateHandler, evseID, chargingConnector, availability.ConnectorStatusAvailable)
	// Send transaction end data
	sampledValue.Context = types.ReadingContextTransactionEnd
	sampledValue.Value = stateHandler.meterValue
	tx.StoppedReason = transactions.ReasonEVDisconnected
	txEventResp, err = chargingStation.TransactionEvent(ctx, transactions.TransactionEventEnded, types.Now(), transactions.TriggerReasonEVCommunicationLost, evse.nextSequence(), tx, func(request *transactions.TransactionEventRequest) {
		request.Evse = &evseReq
		request.IDToken = &dummyClientIdToken
		request.MeterValue = []types.MeterValue{}
	})
	checkError(err)
	logDefault(txEventResp.GetFeatureName()).Infof("transaction %v stopped", tx.TransactionID)

	// Wait for some time ...
	time.Sleep(2 * time.Second)
	// End simulation

	span.AddEvent("simulation ended")
}

// Start function
func main() {
	ctx := context.Background()
	cfg := LoadOTelConfigFromEnv()

	if cfg.EnableTracing {
		shutdownTracing := setupTracing(ctx, cfg.EnableTracing, &cfg)
		defer func(ctx context.Context) {
			if err := shutdownTracing(ctx); err != nil {
				log.Error(err)
			}
		}(ctx)
		log.Info("tracing enabled")
	}

	// Load other config
	id, ok := os.LookupEnv(envVarClientID)
	if !ok {
		log.Printf("no %v environment variable found, exiting...", envVarClientID)
		return
	}
	csmsUrl, ok := os.LookupEnv(envVarCSMSUrl)
	if !ok {
		log.Printf("no %v environment variable found, exiting...", envVarCSMSUrl)
		return
	}
	// Check if TLS enabled
	t, _ := os.LookupEnv(envVarTls)
	tlsEnabled, _ := strconv.ParseBool(t)
	// Prepare OCPP 2.0.1 charging station (chargingStation variable is defined in handler.go)
	if tlsEnabled {
		chargingStation = setupTlsChargingStation(id)
	} else {
		chargingStation = setupChargingStation(id)
	}
	// Setup some basic state management
	evse := EVSEInfo{
		availability:       availability.OperationalStatusOperative,
		currentTransaction: "",
		currentReservation: 0,
		connectors: map[int]ConnectorInfo{
			0: {
				status:       availability.ConnectorStatusAvailable,
				availability: availability.OperationalStatusOperative,
				typ:          reservation.ConnectorTypeCType2,
			},
		},
		seqNo: 0,
	}
	handler := &ChargingStationHandler{
		model:                "model1",
		vendor:               "vendor1",
		availability:         availability.OperationalStatusOperative,
		evse:                 map[int]*EVSEInfo{1: &evse},
		localAuthList:        []localauth.AuthorizationData{},
		localAuthListVersion: 0,
		monitoringLevel:      0,
		meterValue:           0,
	}
	// Support callbacks for all OCPP 2.0.1 profiles
	chargingStation.SetAvailabilityHandler(handler)
	chargingStation.SetAuthorizationHandler(handler)
	chargingStation.SetDataHandler(handler)
	chargingStation.SetDiagnosticsHandler(handler)
	chargingStation.SetDisplayHandler(handler)
	chargingStation.SetFirmwareHandler(handler)
	chargingStation.SetISO15118Handler(handler)
	chargingStation.SetLocalAuthListHandler(handler)
	chargingStation.SetProvisioningHandler(handler)
	chargingStation.SetRemoteControlHandler(handler)
	chargingStation.SetReservationHandler(handler)
	chargingStation.SetSmartChargingHandler(handler)
	chargingStation.SetTariffCostHandler(handler)
	chargingStation.SetTransactionsHandler(handler)
	ocppj.SetLogger(log)

	// Connects to central system
	if err := chargingStation.Start(csmsUrl); err != nil {
		log.Error(err)
	} else {
		log.Infof("connected to CSMS at %v", csmsUrl)
		exampleRoutine(ctx, chargingStation, handler)
		// Disconnect
		chargingStation.Stop()
		log.Infof("disconnected from CSMS")
	}
}

func init() {
	log = logrus.New()
	log.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
	log.SetLevel(logrus.InfoLevel)
}

// Utility functions
func logDefault(feature string) *logrus.Entry {
	return log.WithField("message", feature)
}
