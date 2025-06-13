package main

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// 设置上下文和超时时间
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 连接 MongoDB
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		panic(err)
	}
	defer client.Disconnect(ctx)

	// 选择数据库和集合
	collection := client.Database("pygogo").Collection("submissions")

	// 查询所有文档
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		panic(err)
	}
	defer cursor.Close(ctx)

	// 遍历结果
	fmt.Println("所有提交记录：")
	for cursor.Next(ctx) {
		var result map[string]interface{}
		if err := cursor.Decode(&result); err != nil {
			fmt.Println("解码失败：", err)
			continue
		}
		fmt.Println(result)
	}

	if err := cursor.Err(); err != nil {
		fmt.Println("游标出错：", err)
	}
}
