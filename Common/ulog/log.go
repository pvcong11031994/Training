package ulog

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"producer/common/file"
	"producer/common/uerror"
	"strings"
	"time"

	eParser "github.com/go-errors/errors"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	logNameRequest         = "requests.log"
	logNameWarning         = "warning.log"
	logNameError           = "error.log"
	logNameInfo            = "info.log"
	logNameDebug           = "debug.log"
	logNamePanic           = "panic.log"
	defaultLogDir          = "./log"
	defaultMaxAge          = 15 // 15 days
	defaultMaxBackupFiles  = 15 // 15 files
	defaultMaxBackupSize   = 30 // 30 MB
	omittedSensitiveHeader = "hidden"
	timeFormat             = "2006/01/02 - 15:04:05"
)

type LogConfig struct {
	MaxSizeInMB   int
	MaxAgeDay     int
	MaxBackupFile int
	Debug         bool
}

type ErrLog struct {
	Link      string
	GroupPath string
	Method    string
	Latency   time.Duration
	Error     error
	Status    int
	Client    int
	UserID    int
	Header    http.Header
	Body      string
}

func (l Ulogger) cleanSensitiveHeader(header http.Header) http.Header {
	for _, sensitiveHeader := range l.sensitiveHeaders {
		if existingHeader := header.Get(sensitiveHeader); existingHeader != "" {
			header.Set(sensitiveHeader, omittedSensitiveHeader)
		}
	}
	return header
}

type Ulogger struct {
	LogPath          string
	Request          *lumberjack.Logger
	Warning          *logrus.Logger
	Error            *logrus.Logger
	Info             *logrus.Logger
	Debug            *logrus.Logger
	Panic            *logrus.Logger
	init             bool
	debugMode        bool
	sensitiveHeaders []string
}

var (
	defaultLogger Ulogger
)

func Logger() Ulogger {
	if !defaultLogger.init {
		InitDefaultLogger(defaultLogDir, false, []string{})
		log.Println("[WARNING] log-path is not set, default path './log' will be used")
	}
	return defaultLogger
}

type Fields logrus.Fields

type Http struct {
	Link      string        `json:"link"`
	GroupPath string        `json:"groupPath"`
	Method    string        `json:"method"`
	Latency   time.Duration `json:"latency,omitempty"`
	Status    int           `json:"status"`
	Client    int           `json:"client,omitempty"`
	UserID    int           `json:"userId,omitempty"`
	Header    http.Header   `json:"header,omitempty"`
	Body      string        `json:"body,omitempty"`
	Err       error         `json:"err,omitempty"`
	LatencyMs float64       `json:"latencyMs"`
	Time      string        `json:"time"`
	IP        string        `json:"ip"`
}

func (v Http) JsonMarshal() (string, error) {
	result, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(result) + "\n", nil
}

func (v Http) Fields() Fields {
	ret := Fields{
		"link":    v.Link,
		"method":  v.Method,
		"latency": v.Latency,
		"status":  v.Status,
		"client":  v.Client,
		"userId":  v.UserID,
		"header":  v.Header,
		"body":    v.Body,
	}
	if v.Err != nil {
		ret["error"] = v.Err.Error()
	}
	return ret
}

func InitDefaultLogger(logPath string, debugMode bool, sensitiveHeaders []string) {
	err := file.CreateDir(logPath)
	if err != nil {
		panic(err)
	}

	defaultConfig := LogConfig{
		MaxSizeInMB:   defaultMaxBackupSize,
		MaxAgeDay:     defaultMaxAge,
		MaxBackupFile: defaultMaxBackupFiles,
		Debug:         debugMode,
	}

	defaultLogger = Ulogger{
		LogPath:          logPath,
		init:             true,
		Request:          newLogWriter(path.Join(logPath, logNameRequest), defaultConfig),
		Warning:          newLogger(path.Join(logPath, logNameWarning), logrus.WarnLevel, defaultConfig),
		Error:            newLogger(path.Join(logPath, logNameError), logrus.ErrorLevel, defaultConfig),
		Info:             newLogger(path.Join(logPath, logNameInfo), logrus.InfoLevel, defaultConfig),
		Debug:            newLogger(path.Join(logPath, logNameDebug), logrus.DebugLevel, defaultConfig),
		Panic:            newLogger(path.Join(logPath, logNamePanic), logrus.PanicLevel, defaultConfig),
		debugMode:        defaultConfig.Debug,
		sensitiveHeaders: sensitiveHeaders,
	}
}

// Create logger with path <prefix_name>_<error/info/warning>.log
func NewLogger(logPath, prefixName string, config LogConfig) Ulogger {
	if logPath == "" {
		logPath = defaultLogger.LogPath
	}
	requests := []string{logNameRequest}
	warnings := []string{logNameWarning}
	errors := []string{logNameError}
	infos := []string{logNameInfo}
	debugs := []string{logNameDebug}
	panics := []string{logNamePanic}
	if len(prefixName) > 0 {
		requests = append([]string{prefixName}, requests...)
		warnings = append([]string{prefixName}, warnings...)
		errors = append([]string{prefixName}, errors...)
		infos = append([]string{prefixName}, infos...)
		debugs = append([]string{prefixName}, debugs...)
		panics = append([]string{prefixName}, panics...)
	}

	return Ulogger{
		LogPath:   logPath,
		init:      true,
		Request:   newLogWriter(path.Join(logPath, strings.Join(requests, "_")), config),
		Warning:   newLogger(path.Join(logPath, strings.Join(warnings, "_")), logrus.WarnLevel, config),
		Error:     newLogger(path.Join(logPath, strings.Join(errors, "_")), logrus.ErrorLevel, config),
		Info:      newLogger(path.Join(logPath, strings.Join(infos, "_")), logrus.InfoLevel, config),
		Debug:     newLogger(path.Join(logPath, strings.Join(debugs, "_")), logrus.DebugLevel, config),
		Panic:     newLogger(path.Join(logPath, strings.Join(panics, "_")), logrus.PanicLevel, config),
		debugMode: config.Debug,
	}
}

func HookDefaultLogerWithFluent(fluent Fluent) {
	hook, err := newHookFluent(fluent)
	if err != nil {
		panic(err)
	}
	defaultLogger.AddHook(hook)
}

func HookLoggerWithFluent(log *Ulogger, fluent Fluent) {
	if log == nil {
		panic(fmt.Errorf("logger: nil"))
	}
	hook, err := newHookFluent(fluent)
	if err != nil {
		panic(err)
	}
	log.AddHook(hook)
}

func newLogger(fileName string, level logrus.Level, config LogConfig) *logrus.Logger {
	return &logrus.Logger{
		Out:       newLogWriter(fileName, config),
		Formatter: new(logrus.JSONFormatter),
		Level:     level,
		Hooks:     make(logrus.LevelHooks),
	}
}

func newLogWriter(filename string, config LogConfig) *lumberjack.Logger {
	return &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    config.MaxSizeInMB, // megabytes
		MaxBackups: config.MaxBackupFile,
		MaxAge:     config.MaxAgeDay, // days
	}
}

func (l *Ulogger) AddHook(hook logrus.Hook) {
	if l.Error != nil {
		l.Error.Hooks.Add(hook)
	}
	if l.Warning != nil {
		l.Warning.Hooks.Add(hook)
	}
	if l.Info != nil {
		l.Info.Hooks.Add(hook)
	}
	if l.Debug != nil {
		l.Debug.Hooks.Add(hook)
	}
}

func (l Ulogger) LogPanic(errMsg string, fields Fields) {
	l.Panic.WithFields(logrus.Fields(fields)).Panic(errMsg)
}

func (l Ulogger) LogErrorManual(err error, msg string) {
	err = GetErrorStackTrace(err)
	stackTrace := fmt.Sprint(err)
	l.Error.WithFields(logrus.Fields{
		"time":       time.Now(),
		"stackTrace": stackTrace,
	}).Error(msg)
}

func (l Ulogger) LogError(msg string, fields Fields) {
	l.Error.WithFields(logrus.Fields(fields)).
		Error(msg)
}

func (l Ulogger) LogInfo(msg string, fields Fields) {
	l.Info.WithFields(logrus.Fields(fields)).
		Info(msg)
}

func (l Ulogger) LogWarn(msg string, fields Fields) {
	l.Warning.WithFields(logrus.Fields(fields)).
		Warn(msg)
}

func (l Ulogger) LogDebug(msg string, fields Fields) {
	l.Debug.WithFields(logrus.Fields(fields)).
		Debug(msg)
}

func (l Ulogger) LogErrorObjectManual(err error, msg string, data interface{}) {
	dataStr := fmt.Sprintf("%+v", data)
	err = GetErrorStackTrace(err)
	stackTrace := fmt.Sprint(err)
	l.Error.WithFields(logrus.Fields{
		"time":       time.Now(),
		"data":       dataStr,
		"stackTrace": stackTrace,
	}).Error(msg)
}

func (l Ulogger) LogDebugObject(mess string, data interface{}) {
	if !l.debugMode {
		return
	}
	dataStr := fmt.Sprintf("%+v", data)
	l.Debug.WithFields(logrus.Fields{
		"time": time.Now(),
		"data": dataStr,
	}).Debug(mess)

}

func (l Ulogger) LogInfoObject(mess string, data interface{}) {
	dataStr := fmt.Sprintf("%+v", data)
	l.Info.WithFields(logrus.Fields{
		"time": time.Now(),
		"data": dataStr,
	}).Info(mess)
}

func (l Ulogger) LogInfoStat(duration time.Duration, msg string) {
	l.Info.WithFields(logrus.Fields{
		"time":     time.Now(),
		"duration": duration,
	}).Info(msg)
}

func (l Ulogger) LogDebugStat(duration time.Duration, msg string) {
	if !l.debugMode {
		return
	}
	l.Debug.WithFields(logrus.Fields{
		"time":     time.Now(),
		"duration": duration,
	}).Debug(msg)
}

func (l Ulogger) LogErrorStat(duration time.Duration, msg string, err error) {

	l.Error.WithFields(logrus.Fields{
		"time":     time.Now(),
		"duration": duration,
		"error":    err,
	}).Error(msg)
}

func (l Ulogger) LogWarning(err error, data interface{}, msg string) {
	err = GetErrorStackTrace(err)
	dataStr := fmt.Sprintf("%+v", data)
	stackTrace := fmt.Sprint(uerror.StackTrace(err))
	l.Warning.WithFields(logrus.Fields{
		"data":       dataStr,
		"stackTrace": stackTrace,
	}).Warn(msg)
}

func (l Ulogger) LogRequestObject(req *http.Request, body string, input interface{}, data interface{}, msg string) {
	safeHeader := l.cleanSensitiveHeader(req.Header)

	l.Info.WithFields(logrus.Fields{
		"body":   body,
		"input":  input,
		"res":    data,
		"link":   req.RequestURI,
		"method": req.Method,
		"client": req.Header.Get("client"),
		"header": safeHeader,
	}).Info(msg)
}

func getLogDataMessageAmqp(msg *amqp.Delivery, funcName string, latency time.Duration, err error) logrus.Fields {
	logData := logrus.Fields{
		"funcName": funcName,
		"latency":  latency,
	}
	if err != nil {
		logData["error"] = uerror.StackTrace(err)
	}
	if msg != nil {
		logData["msgBody"] = string(msg.Body)
		logData["msgType"] = msg.Type
		logData["msgContentType"] = msg.ContentType
	}
	return logData
}

func (l Ulogger) LogErrorMessageAmqp(msg *amqp.Delivery, funcName string, note string, latency time.Duration, err error) {
	logData := getLogDataMessageAmqp(msg, funcName, latency, err)
	l.Error.WithFields(logData).Error(note)
}

func (l Ulogger) LogInfoMessageAmqp(msg *amqp.Delivery, funcName string, note string, latency time.Duration) {
	logData := getLogDataMessageAmqp(msg, funcName, latency, nil)
	l.Info.WithFields(logData).Info(note)
}

func (l Ulogger) LogWarningMessageAmqp(msg *amqp.Delivery, funcName string, note string, latency time.Duration) {
	logData := getLogDataMessageAmqp(msg, funcName, latency, nil)
	l.Warning.WithFields(logData).Warn(note)
}

func (l Ulogger) LogDebugMessageAmqp(msg *amqp.Delivery, funcName string, note string, latency time.Duration) {
	if !l.debugMode {
		return
	}
	logData := getLogDataMessageAmqp(msg, funcName, latency, nil)
	l.Debug.WithFields(logData).Debug(note)
}

func GetErrorStackTrace(e error) error {
	return errors.New(eParser.Wrap(e, 0).ErrorStack())
}

func BackupBody(req *http.Request) string {
	var bodyBytes []byte
	if req.Body != nil {
		bodyBytes, _ = ioutil.ReadAll(req.Body)
	}
	// Restore the io.ReadCloser to its original state
	req.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	// Use the content
	bodyString := string(bodyBytes)
	if len(bodyString) > 2000 {
		bodyString = bodyString[:2000]
	}
	return bodyString
}

func (l Ulogger) LogErrorHttp(errLog ErrLog) {
	errMess := fmt.Sprint(errLog.Error)

	safeHeader := l.cleanSensitiveHeader(errLog.Header)

	l.Error.WithFields(logrus.Fields{
		"time":      time.Now(),
		"status":    errLog.Status,
		"method":    errLog.Method,
		"header":    safeHeader,
		"link":      errLog.Link,
		"groupPath": errLog.GroupPath,
		"userId":    errLog.UserID,
		"client":    errLog.Client,
		"body":      errLog.Body,
	}).Error(errMess)
}
