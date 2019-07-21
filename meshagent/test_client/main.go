package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/tormaroe/eightlegs-project/meshagent/api"
	"google.golang.org/grpc"
)

func sendStatus(client api.MeshAgentClient, status *api.MeshServiceStatus) {
	log.Println("Sending status..")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	res, err := client.MyStatusUpdate(ctx, status)
	if err != nil {
		log.Fatalf("%v.MyStatusUpdate(_) = _, %v: ", client, err)
	}
	fmt.Println(res)
}

func main() {
	fmt.Println("MeshAgent test client")

	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())

	conn, err := grpc.Dial("127.0.0.1:50710", opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()

	client := api.NewMeshAgentClient(conn)

	sendStatus(client, &api.MeshServiceStatus{
		ServiceUuid: "5ab5d266-b9b4-4a10-bb23-3acec1a80092",
		ServiceType: "TestClient",
	})
}
