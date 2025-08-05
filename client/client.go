package main

import (
	// ...
	"context"
	"fmt"
	"log"

	pb "github.com/sinfirst/URL-Cutter/proto/url_cutter"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// устанавливаем соединение с сервером
	conn, err := grpc.NewClient(":3200", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Println("все печально", err)
	}
	defer conn.Close()

	c := pb.NewURLCutterClient(conn)

	// функция, в которой будем отправлять сообщения
	TestUsers(c)
}

func TestUsers(c pb.URLCutterClient) {
	resp, err := c.PostHandler(context.Background(), &pb.PostHandlerRequest{Url: "132112313fdsfsda"})
	if err != nil {
		log.Fatal(err)
	}
	if resp.Error != "" {
		fmt.Println(resp.Error)
	}
	fmt.Println(resp.ShortUrl)
}
