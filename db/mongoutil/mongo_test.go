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

package mongoutil

import (
	"context"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

func Test_connectWithRetry(t *testing.T) {
	var config = &Config{
		Address:                     []string{"docdb.cluster-czw8k2aoeyoa.us-east-2.docdb.amazonaws.com:27017", "docdb.cluster-ro-czw8k2aoeyoa.us-east-2.docdb.amazonaws.com:27017"},
		Database:                    "nextim",
		Username:                    "rw-nextim",
		Password:                    "LMnd7jKKsd9nndJHBzB",
		ReplicaSet:                  "rs0",
		ReadPreference:              ReadPreferenceSecondaryPreferred,
		TLSEnabled:                  true,
		TlsCAFile:                   "global-bundle.pem",
		TlsAllowInvalidCertificates: false,
		MaxPoolSize:                 100,
		MinPoolSize:                 1,
		MaxRetry:                    10,
		RetryWrites:                 false,
		RetryReads:                  false,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := NewMongoDB(ctx, config)
	if err != nil {
		t.Fatalf("connect failed: %v", err)
	}
	defer client.db.Client().Disconnect(ctx)
	t.Log("success connect MongoDB")

	collection := client.db.Collection("test_collection")
	if collection == nil {

		if err = client.db.CreateCollection(ctx, "test_collection"); err != nil {
			t.Logf("collect exsit: %v", err)
		}
	}

	testDoc := bson.M{
		"name":    "test_user",
		"value":   42,
		"created": time.Now(),
	}
	insertRes, err := collection.InsertOne(ctx, testDoc)
	if err != nil {
		t.Fatalf("insert failed: %v", err)
	}
	t.Logf("insert success，ID: %v", insertRes.InsertedID)

	var result bson.M
	filter := bson.M{"name": "test_user"}
	err = collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		t.Fatalf("select failed: %v", err)
	}
	t.Logf("select result: %+v", result)

	if result["value"] != int32(42) {
		t.Errorf("not same，want 42, ext: %v", result["value"])
	}

	if _, err = collection.DeleteMany(ctx, filter); err != nil {
		t.Logf("delete failed: %v", err)
	}
}

//func Test_connectWithRetry(t *testing.T) {
//	type args struct {
//		ctx         context.Context
//		clientOpts  *options.ClientOptions
//		maxRetry    int
//		connTimeout time.Duration
//	}
//	tests := []struct {
//		name    string
//		args    args
//		want    *mongo.Client
//		wantErr bool
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			got, err := connectWithRetry(tt.args.ctx, tt.args.clientOpts, tt.args.maxRetry, tt.args.connTimeout)
//			if (err != nil) != tt.wantErr {
//				t.Errorf("connectWithRetry() error = %v, wantErr %v", err, tt.wantErr)
//				return
//			}
//			if !reflect.DeepEqual(got, tt.want) {
//				t.Errorf("connectWithRetry() = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}
//
//func TestCheckMongo(t *testing.T) {
//	type args struct {
//		ctx    context.Context
//		config *Config
//	}
//	tests := []struct {
//		name    string
//		args    args
//		wantErr bool
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			if err := CheckMongo(tt.args.ctx, tt.args.config); (err != nil) != tt.wantErr {
//				t.Errorf("CheckMongo() error = %v, wantErr %v", err, tt.wantErr)
//			}
//		})
//	}
//}
