package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gogo/protobuf/types"
	"google.golang.org/grpc"

	"github.com/tarndt/wasmws"
)

func main() {

	//App context setup
	appCtx, appCancel := context.WithCancel(context.Background())
	defer appCancel()

	//Setup HTTP / Websocket server
	router := http.NewServeMux()
	wsl := wasmws.NewWebSocketListener(appCtx)
	router.HandleFunc("/grpc-proxy", wsl.ServeHTTP)
	router.Handle("/", http.FileServer(http.Dir(".")))
	log.Println("starting http/ws server at 0.0.0.0:8080")
	httpServer := &http.Server{Addr: ":8080", Handler: router}
	//Run HTTP server
	go func() {
		defer appCancel()
		log.Printf("ERROR: HTTP Listen and Server failed; Details: %s", httpServer.ListenAndServe())
	}()

	grpcServer := grpc.NewServer()
	RegisterSimpleServiceServer(grpcServer, new(streamServer))
	//Run gRPC server
	go func() {
		defer appCancel()

		if err := grpcServer.Serve(wsl); err != nil {
			log.Printf("ERROR: Failed to serve gRPC connections; Details: %s", err)
		}
	}()
	//Run gRPC server
	go func() {
		defer appCancel()

		if err := grpcServer.Serve(wsl); err != nil {
			log.Printf("ERROR: Failed to serve gRPC connections; Details: %s", err)
		}
	}()

	//Handle signals
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		log.Printf("INFO: Received shutdown signal: %s", <-sigs)
		appCancel()
	}()

	//Shutdown
	<-appCtx.Done()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), time.Second*2)
	defer shutdownCancel()

	grpcShutdown := make(chan struct{}, 1)
	go func() {
		grpcServer.GracefulStop()
		grpcShutdown <- struct{}{}
	}()

	httpServer.Shutdown(shutdownCtx)
	select {
	case <-grpcShutdown:
	case <-shutdownCtx.Done():
		grpcServer.Stop()
	}
}

type streamServer struct{}

func (*streamServer) ReapServerStream(req *types.Empty, srv SimpleService_ReapServerStreamServer) error {

	for i := 0; i < 500; i++ {
		err := srv.Send(&ReplyMessage{
			Timestamp: types.TimestampNow(),
			Message:   fmt.Sprintf("message%d", i),
		})
		if err != nil {
			return fmt.Errorf("Send the %d message error: %s", i, err)
		}
	}

	return nil
}
