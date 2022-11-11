//go:build unit
// +build unit

/*
 * Copyright 2019 Eric Evans <eevans@wikimedia.org>, and Wikimedia
 * Foundation
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"testing"
)

type mockIOWriter struct {
	data []byte
}

func (m *mockIOWriter) Write(data []byte) (n int, err error) {
	m.data = data
	return len(m.data), nil
}

func TestLogger(t *testing.T) {
	setUp := func(level int) (*mockIOWriter, *Logger) {
		writer := &mockIOWriter{}
		logger, _ := NewLogger(writer, "logtest", LevelString(level))
		return writer, logger
	}

	getLogMessage := func(writer *mockIOWriter) (LogMessage, error) {
		var result LogMessage
		err := json.Unmarshal(writer.data, &result)
		return result, err
	}

	t.Run("Simple logging", func(t *testing.T) {
		testCases := []struct {
			format string
			arg    string
			level  int
		}{
			{"Debug %s", "your bugs", LogDebug},
			{"Info %s", "wars", LogInfo},
			{"Consider yourself %s", "warned", LogWarning},
			{"Errors are %s", "bad", LogError},
			{"Fatal %s", "attraction", LogFatal},
		}
		for _, tcase := range testCases {
			t.Run(LevelString(tcase.level), func(t *testing.T) {
				writer, logger := setUp(LogDebug)

				switch tcase.level {
				case LogDebug:
					logger.Debug(tcase.format, tcase.arg)
				case LogInfo:
					logger.Info(tcase.format, tcase.arg)
				case LogWarning:
					logger.Warning(tcase.format, tcase.arg)
				case LogError:
					logger.Error(tcase.format, tcase.arg)
				case LogFatal:
					logger.Fatal(tcase.format, tcase.arg)
				default:
					t.Fatalf("Testcase has invalid level!")
				}

				r, err := getLogMessage(writer)
				if err != nil {
					t.Fatalf("Unable to deserialize JSON log message: %s", err)
				}

				AssertEquals(t, fmt.Sprintf(tcase.format, tcase.arg), r.Msg, "Wrong message string attribute")
				AssertEquals(t, LevelString(tcase.level), r.Level, "Wrong log level attribute")
				AssertEquals(t, "logtest", r.Appname, "Wrong appname attribute")
			})
		}
	})

	// Logger is configured for INFO and above (DEBUG should be ignored)
	t.Run("Filtered", func(t *testing.T) {
		writer, logger := setUp(LogInfo)
		logger.Debug("Noisy log message")
		AssertEquals(t, 0, len(writer.data), "Unexpected log output")
	})

	t.Run("Scoped", func(t *testing.T) {
		writer, logger := setUp(LogInfo)
		logger.RequestID("0000000a-000a-000a-000a-00000000000a").Log(LogWarning, "Consider yourself %s", "warned")

		res, err := getLogMessage(writer)
		if err != nil {
			t.Fatalf("Unable to deserialize JSON log message: %s", err)
		}

		AssertEquals(t, "Consider yourself warned", res.Msg, "Wrong message string attribute")
		AssertEquals(t, LevelString(LogWarning), res.Level, "Wrong log level attribute")
		AssertEquals(t, "logtest", res.Appname, "Wrong appname attribute")
		AssertEquals(t, "0000000a-000a-000a-000a-00000000000a", res.RequestID, "Wrong request_id attribute")
	})

	t.Run("Using log module", func(t *testing.T) {
		writer, logger := setUp(LogInfo)
		log.SetFlags(0)
		log.SetOutput(logger)
		log.Println("Sent via log module")

		res, err := getLogMessage(writer)
		if err != nil {
			t.Fatalf("Unable to deserialize JSON log message: %s", err)
		}

		AssertEquals(t, "Sent via log module", res.Msg, "Wrong message string attribute")
		AssertEquals(t, LevelString(LogWarning), res.Level, "Wrong log level attribute")
	})
}
