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

	// 1. 连接数据库
	client, err := NewMongoDB(ctx, config)
	if err != nil {
		t.Fatalf("连接失败: %v", err)
	}
	defer client.db.Client().Disconnect(ctx)
	t.Log("成功连接到 MongoDB")

	// 2. 获取集合对象
	collection := client.db.Collection("test_collection")
	if collection == nil {
		// 3. 创建集合（可选，MongoDB 会在首次插入时自动创建）
		if err = client.db.CreateCollection(ctx, "test_collection"); err != nil {
			t.Logf("集合可能已存在: %v", err) // 忽略已存在错误
		}
	}

	//4. 写入测试数据
	testDoc := bson.M{
		"name":    "test_user",
		"value":   42,
		"created": time.Now(),
	}
	insertRes, err := collection.InsertOne(ctx, testDoc)
	if err != nil {
		t.Fatalf("插入失败: %v", err)
	}
	t.Logf("插入成功，ID: %v", insertRes.InsertedID)

	// 5. 读取测试数据
	var result bson.M
	filter := bson.M{"name": "test_user"}
	err = collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		t.Fatalf("查询失败: %v", err)
	}
	t.Logf("查询结果: %+v", result)

	// 验证数据一致性
	if result["value"] != int32(42) {
		t.Errorf("数据不一致，期望 42，实际 %v", result["value"])
	}

	// 6. 清理测试数据（可选）
	if _, err = collection.DeleteMany(ctx, filter); err != nil {
		t.Logf("清理失败: %v", err)
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
