// Copyright © 2024 OpenIM open source community. All rights reserved.
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

package kafka

import "time"

type TLSConfig struct {
	EnableTLS          bool   `yaml:"enableTLS"`
	CACrt              string `yaml:"caCrt"`
	ClientCrt          string `yaml:"clientCrt"`
	ClientKey          string `yaml:"clientKey"`
	ClientKeyPwd       string `yaml:"clientKeyPwd"`
	InsecureSkipVerify bool   `yaml:"insecureSkipVerify"`
}

type Config struct {
	Username     string         `yaml:"username"`
	Password     string         `yaml:"password"`
	ProducerAck  string         `yaml:"producerAck"`
	CompressType string         `yaml:"compressType"`
	Addr         []string       `yaml:"addr"`
	TLS          TLSConfig      `yaml:"tls"`
	ClientID     string         `yaml:"clientID"`
	Metadata     ConfigMetadata `yaml:"metadata"`
}

type ConfigMetadata struct {
	RefreshFrequency time.Duration `yaml:"refreshFrequency"`
}
