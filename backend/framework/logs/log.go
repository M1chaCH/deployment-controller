package logs

import (
	"github.com/gin-gonic/gin"
	"go.elastic.co/apm/v2"
	"os"

	"github.com/M1chaCH/deployment-controller/framework/config"
	"github.com/sirupsen/logrus"
	"go.elastic.co/ecslogrus"
)

// This is my usage of logging
// currently the APM integration is required and must be configured
// also currently the logrus usage can not be changed
// everywhere in here the gin.Context must remain optional

var logger *logrus.Logger
var serviceName string
var serviceEnvironment string

func InitLogging() {
	cnf := config.Config()

	serviceName = cnf.APM.ServiceName
	serviceEnvironment = cnf.APM.Environment

	if &cnf.APM == nil || cnf.APM.ApiKey == "" {
		panic("APM API key not set, cannot initialize APM logging")
	}

	logger = logrus.New()
	logger.SetFormatter(&ecslogrus.Formatter{})
	logger.SetReportCaller(true)

	if cnf.Log.Level > -1 && cnf.Log.Level < 6 {
		logger.SetLevel(logrus.Level(cnf.Log.Level))
	} else {
		logger.SetLevel(logrus.DebugLevel)
	}

	if cnf.Log.FileName != "" {
		logFile, err := os.OpenFile(cnf.Log.FileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
		if err != nil {
			panic("could not setup logger at " + cnf.Log.FileName + ": " + err.Error() + "")
		}

		logger.SetOutput(logFile)
	}

	Info(nil, "logrus logger initialized with ecs formatter")
}

func Debug(c *gin.Context, message string, args ...interface{}) {
	fields := getApmData(c)
	logger.WithFields(fields).Debugf(message, args...)
}

func Info(c *gin.Context, message string, args ...interface{}) {
	fields := getApmData(c)
	logger.WithFields(fields).Infof(message, args...)
}

func Warn(c *gin.Context, message string, args ...interface{}) {
	fields := getApmData(c)
	logger.WithFields(fields).Warningf(message, args...)
}

func Error(c *gin.Context, message string, args ...interface{}) {
	fields := getApmData(c)
	logger.WithFields(fields).Errorf(message, args...)
}

func Panic(c *gin.Context, message string, args ...interface{}) {
	fields := getApmData(c)
	logger.WithFields(fields).Panicf(message, args...)
}

func getApmData(c *gin.Context) logrus.Fields {
	traceId, transactionId := "", ""

	if c != nil {
		traceContext := apm.TransactionFromContext(c.Request.Context()).TraceContext()
		traceId = traceContext.Trace.String()
		transactionId = traceContext.Span.String()
	}

	fields := logrus.Fields{
		"trace.id":                   traceId,
		"transaction.id":             transactionId,
		"service.name":               serviceName,
		"service.serviceEnvironment": serviceEnvironment,
	}

	return fields
}

type ApmLabels map[string]interface{}

func AddApmLabels(c *gin.Context, labels map[string]interface{}) {
	apmTx := apm.TransactionFromContext(c.Request.Context())
	for k, v := range labels {
		apmTx.Context.SetLabel(k, v)
	}
}
