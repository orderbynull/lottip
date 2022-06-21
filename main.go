package main

import (
	"flag"
	"fmt"
	"github.com/kjk/dailyrotate"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"io"
	"lottip/chat"
	"os"
	"strings"
	"syscall"
	"time"
)

var (
	proxyAddr          = flag.String("proxy", "127.0.0.1:4041", "Proxy <host>:<port>")
	logRequests        = flag.Bool("log-requests", false, "Enable logging of requests")
	logResponses       = flag.Bool("log-responses", false, "Enable logging of responses")
	logResponsePackets = flag.Bool("log-response-packets", false, "Enable logging of response packets")
	logAll             = flag.Bool("log-all", true, "Enable logging of requests, responses, and other events")
	logToConsole       = flag.Bool("enable-console-logging", false, "Enable logging to console")
	logToFile          = flag.Bool("enable-file-logging", true, "Enable logging to console")
	logDirectory       = flag.String("log-directory", "./logs", "Set the query log directory")
	queryLogFile       = flag.String("query-log-file", "queries.log", "Set the query log file name")
	logFile            = flag.String("log-file", "logfile.log", "Set the query log file name")
	mysqlAddr          = flag.String("mysql", "127.0.0.1:3306", "MySQL <host>:<port>")
	guiAddr            = flag.String("gui-addr", "127.0.0.1:9999", "Web UI <host>:<port>")
	guiEnabled         = flag.Bool("gui-enabled", false, "Enable the web-gui server")
	useLocalUI         = flag.Bool("use-local", false, "Use local UI instead of embed")
	mysqlDsn           = flag.String("mysql-dsn", "", "MySQL DSN for query execution capabilities")

	queryLogger zerolog.Logger
)

func appReadyInfo(appReadyChan chan bool) {
	<-appReadyChan
	time.Sleep(1 * time.Second)
	log.Info().Msgf("Forwarding queries from `%s` to `%s`", *proxyAddr, *mysqlAddr)
	log.Info().Msgf("Web gui available at `http://%s`", *guiAddr)
}

func timespecToTime(ts syscall.Timespec) time.Time {
	return time.Unix(int64(ts.Sec), int64(ts.Nsec))
}

func newRollingFile(directory string, filename string) io.Writer {
	if err := os.MkdirAll(directory, 0744); err != nil {
		log.Error().Err(err).Str("path", directory).Msg("can't create log directory")
		return nil
	}

	logFileWriter, err := dailyrotate.NewFileWithPathGenerator(func(time time.Time) string {
		return directory + "/" + filename
	}, func(filename string, didRotate bool) {
		if didRotate {
			// Then rename the file
			finfo, _ := os.Stat(filename)
			stat_t := finfo.Sys().(*syscall.Stat_t)
			timeFormatString := ".2006-01-02"
			rolledName := directory + "/" + filename + timespecToTime(stat_t.Birthtimespec).Format(timeFormatString)
			os.Rename(filename, rolledName)
		}
	})

	if err != nil {
		log.Error().Err(err).Str("file", directory+"/"+filename).Msg("Can't create log file")
		return nil
	}

	return logFileWriter
}

func main() {
	flag.Parse()

	if *logAll || *logRequests || *logResponses || *logResponsePackets {
		var queryWriters []io.Writer
		if *logToConsole {
			output := zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "2006-01-02 15:04:05.000"}
			output.FormatLevel = func(i interface{}) string {
				return strings.ToUpper(fmt.Sprintf("%-6s", i))
			}
			queryWriters = append(queryWriters, output)
		}

		if *logToFile {
			queryWriters = append(queryWriters, newRollingFile(*logDirectory, *queryLogFile))
		}

		queryLogMultiWriter := io.MultiWriter(queryWriters...)

		queryLogger = zerolog.New(queryLogMultiWriter).With().Timestamp().Logger()
	}

	zerolog.TimeFieldFormat = "2006-01-02 15:04:05.000"
	zerolog.TimestampFieldName = "logTimestamp"

	var writers []io.Writer
	if *logToConsole {
		output := zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "2006-01-02 15:04:05.000"}
		output.FormatLevel = func(i interface{}) string {
			return strings.ToUpper(fmt.Sprintf("%-6s", i))
		}
		writers = append(writers, output)
	}

	if *logToFile {
		writers = append(writers, newRollingFile(*logDirectory, *logFile))
	}

	logMultiWriter := io.MultiWriter(writers...)

	log.Logger = zerolog.New(logMultiWriter).With().Timestamp().Logger()

	cmdChan := make(chan chat.Cmd)
	cmdResultChan := make(chan chat.CmdResult)
	connStateChan := make(chan chat.ConnState)
	appReadyChan := make(chan bool)

	hub := chat.NewHub(cmdChan, cmdResultChan, connStateChan)

	go hub.Run()
	if guiEnabled != nil && *guiEnabled {
		go runHttpServer(hub)
	}
	go appReadyInfo(appReadyChan)

	p := MySQLProxyServer{cmdChan, cmdResultChan, connStateChan, appReadyChan, *mysqlAddr, *proxyAddr}
	p.run()
}
