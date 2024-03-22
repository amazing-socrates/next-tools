// Copyright © 2023 OpenIM. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package component

import (
	"errors"
	"fmt"
	"github.com/openimsdk/tools/utils/dataformat"
	"io"
	"log"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/IBM/sarama"
	"github.com/go-zookeeper/zk"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/openimsdk/tools/errs"
)

const (
	// DefaultCfgPath is the default path of the configuration file.
	minioHealthCheckDuration = 1
	mongoConnTimeout         = 5 * time.Second
	MaxRetry                 = 300
)

const (
	colorRed    = 31
	colorGreen  = 32
	colorYellow = 33
)

func exactIP(urlStr string) (string, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return "", errs.Wrap(fmt.Errorf("url parse error, err:%v. url:%s", err, urlStr))
	}
	host, _, err := net.SplitHostPort(u.Host)
	if err != nil {
		host = u.Host
	}
	if strings.HasSuffix(host, ":") {
		host = host[0 : len(host)-1]
	}
	return host, nil
}

// CheckMinio checks the MinIO connection.
func CheckMinio(minioStu *Minio) error {
	if minioStu.Endpoint == "" || minioStu.AccessKeyID == "" || minioStu.SecretAccessKey == "" {
		return errs.Wrap(errors.New("missing configuration for endpoint, accessKeyID, or secretAccessKey"))
	}

	// Parse endpoint URL to determine if SSL is enabled
	minioInfo, err := dataformat.JsonMarshal(minioStu)
	if err != nil {
		return errs.Wrap(errors.New("minioStu Marshal failed"))
	}
	u, err := url.Parse(minioStu.Endpoint)
	if err != nil {
		return errs.Wrap(fmt.Errorf("url parse failed, err:%v. %s", err, string(minioInfo)))
	}
	secure := u.Scheme == "https" || minioStu.UseSSL == "true"

	// Initialize MinIO client
	minioClient, err := minio.New(u.Host, &minio.Options{
		Creds:  credentials.NewStaticV4(minioStu.AccessKeyID, minioStu.SecretAccessKey, ""),
		Secure: secure,
	})
	if err != nil {
		return errs.Wrap(fmt.Errorf("initialize minio client failed, err:%v. %s", err, string(minioInfo)))
	}

	// Perform health check
	cancel, err := minioClient.HealthCheck(time.Duration(minioHealthCheckDuration) * time.Second)
	if err != nil {
		return errs.Wrap(fmt.Errorf("minio client health check failed, err:%v. %s", err, string(minioInfo)))
	}
	defer cancel()

	if minioClient.IsOffline() {
		return errs.Wrap(fmt.Errorf("minio client is offline. %s", string(minioInfo)))
	}

	// Check for localhost in API URL and Minio SignEndpoint
	apiURL, err := exactIP(minioStu.ApiURL)
	if err != nil {
		return err
	}
	signEndPoint, err := exactIP(minioStu.SignEndpoint)
	if err != nil {
		return err
	}
	if apiURL == "127.0.0.1" {
		ErrorPrint(fmt.Sprintf("Warning, minioStu.apiURL(%s) contain 127.0.0.1.", minioStu.ApiURL))
	}
	if signEndPoint == "127.0.0.1" {
		ErrorPrint(fmt.Sprintf("Warning, minioStu.signEndPoint(%s) contain 127.0.0.1.", minioStu.SignEndpoint))
	}
	return nil
}

func CheckZookeeper(zkStu *Zookeeper) error {
	zkStuInfo, err := dataformat.JsonMarshal(zkStu)
	if err != nil {
		return fmt.Errorf("zkStu Marshal failed: %w", err)
	}

	// Temporary disable logging
	originalLogger := log.Default().Writer()
	log.SetOutput(io.Discard)
	defer log.SetOutput(originalLogger) // Ensure logging is restored

	// Connect to Zookeeper
	c, _, err := zk.Connect(zkStu.ZkAddr, time.Second) // Adjust the timeout as necessary
	if err != nil {
		return fmt.Errorf("zk connect failed, err: %v. %s", err, string(zkStuInfo))
	}
	defer c.Close()

	// Check if we can get a connection within 5 seconds
	connected := make(chan bool)
	go func() {
		for {
			if c.State() == zk.StateHasSession {
				connected <- true
				return
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()

	select {
	case <-connected:
		// Connection successful
		return nil
	case <-time.After(5 * time.Second):
		// Connection timed out
		return fmt.Errorf("timeout waiting for Zookeeper connection: %s", string(zkStuInfo))
	}
}

// CheckKafka checks the Kafka connection.
func CheckKafka(kafkaStu *Kafka) (sarama.Client, error) {
	// Configure Kafka client
	cfg := sarama.NewConfig()
	if kafkaStu.Username != "" && kafkaStu.Password != "" {
		cfg.Net.SASL.Enable = true
		cfg.Net.SASL.User = kafkaStu.Username
		cfg.Net.SASL.Password = kafkaStu.Password
	}
	// Additional Kafka setup (e.g., TLS configuration) can be added here
	// kafka.SetupTLSConfig(cfg)

	// Create Kafka client
	kafkaInfo, err := dataformat.JsonMarshal(kafkaStu)
	if err != nil {
		return nil, errs.Wrap(errors.New("kafkaStu Marshal failed"))
	}
	kafkaClient, err := sarama.NewClient(kafkaStu.Addr, cfg)
	if err != nil {
		return nil, errs.Wrap(fmt.Errorf("kafaka connected failed, err:%v. %s", err, string(kafkaInfo)))
	}

	return kafkaClient, nil
}

func colorPrint(colorCode int, format string, a ...any) {
	fmt.Printf("\x1b[%dm%s\x1b[0m\n", colorCode, fmt.Sprintf(format, a...))
}

func colorErrPrint(colorCode int, format string, a ...any) {
	log.Printf("\x1b[%dm%s\x1b[0m\n", colorCode, fmt.Sprintf(format, a...))
}

func ErrorPrint(s string) {
	colorErrPrint(colorRed, "%v", s)
}

func SuccessPrint(s string) {
	colorPrint(colorGreen, "%v", s)
}

func WarningPrint(s string) {
	colorPrint(colorYellow, "Warning: But %v", s)
}

func ErrStr(err error, str string) error {
	return fmt.Errorf("%v;%s", err, str)
}
