package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"sync"

	pb "example.com/grpcbidichatUPGRADE/pkg/chat"
	"google.golang.org/grpc"
)

type chatServer struct {
	pb.UnimplementedChatServiceServer
	mu      sync.Mutex
	clients map[string]pb.ChatService_ChatServer
}

func newChatServer() *chatServer {
	return &chatServer{
		clients: make(map[string]pb.ChatService_ChatServer),
	}
}

func (s *chatServer) Chat(stream pb.ChatService_ChatServer) error {
	var userName string

	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Error receiving: %v", err)
			break
		}

		// Đăng ký lần đầu
		if userName == "" {
			userName = msg.From
			s.mu.Lock()
			s.clients[userName] = stream
			s.mu.Unlock()
			log.Printf("User %s connected", userName)
			continue
		}

		log.Printf("Message from %s to %s: %s", msg.From, msg.To, msg.Message)

		// Gửi tới người nhận
		s.mu.Lock()
		targetStream, ok := s.clients[msg.To]
		s.mu.Unlock()

		if ok {
			err := targetStream.Send(&pb.ChatMessage{
				From:    msg.From,
				To:      msg.To,
				Message: msg.Message,
			})
			if err != nil {
				log.Printf("Send error to %s: %v", msg.To, err)
			}
		} else {
			log.Printf("User %s not found", msg.To)
		}
	}

	// Cleanup khi client rớt
	s.mu.Lock()
	delete(s.clients, userName)
	s.mu.Unlock()
	log.Printf("User %s disconnected", userName)

	return nil
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal(err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterChatServiceServer(grpcServer, newChatServer())
	fmt.Println("🚀 Server listening on :50051")
	grpcServer.Serve(lis)
}
