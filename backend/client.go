package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	pb "forchatic/pkg/chat"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter your name: ")
	user, _ := reader.ReadString('\n')
	user = strings.TrimSpace(user)

	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := pb.NewChatServiceClient(conn)
	stream, err := client.Chat(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	// Đăng ký tên
	stream.Send(&pb.ChatMessage{From: user})

	// Gửi tin nhắn
	go func() {
		for {
			fmt.Print("To (user): ")
			to, _ := reader.ReadString('\n')
			to = strings.TrimSpace(to)

			fmt.Print("Message: ")
			msg, _ := reader.ReadString('\n')
			msg = strings.TrimSpace(msg)

			stream.Send(&pb.ChatMessage{
				From:    user,
				To:      to,
				Message: msg,
			})
		}
	}()

	// Nhận tin nhắn
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("\n💬 %s says: %s\n", in.From, in.Message)
	}
}
