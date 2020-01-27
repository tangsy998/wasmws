//+build wasm,js

package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"syscall/js"
	"time"

	"github.com/gogo/protobuf/types"
	"google.golang.org/grpc"

	"github.com/tarndt/wasmws"
)

var htDoc = js.Global().Get("document")
var nl, elBody js.Value

func main() {
	nl = htDoc.Call("getElementsByTagName", "body")
	if nl.Length() > 0 {
		elBody = nl.Index(0)
	}
	p1 := htDoc.Call("createElement", "p")
	elBody.Call("appendChild", p1)

	elInput := htDoc.Call("createElement", "input")
	elInput.Set("type", "text")
	elInput.Set("value", "demo")
	p1.Call("appendChild", elInput)

	btn := htDoc.Call("createElement", "button")
	btn.Set("type", "button")
	btn.Set("textContent", "Go!")
	p1.Call("appendChild", btn)

	ch := make(chan string, 1)
	btn.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		ch <- elInput.Get("value").String()
		return nil
	}))

	//App context setup
	appCtx, appCancel := context.WithCancel(context.Background())
	defer appCancel()

	//Dial setup
	const dialTO = time.Second
	dialCtx, dialCancel := context.WithTimeout(appCtx, dialTO)
	defer dialCancel()

	//Connect to remote gRPC server
	const websocketURL = "ws://localhost:8080/grpc-proxy"
	conn, err := grpc.DialContext(dialCtx, "passthrough:///"+websocketURL, grpc.WithContextDialer(wasmws.GRPCDialer), grpc.WithDisableRetry(), grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Could not gRPC dial: %s; Details: %s", websocketURL, err)
	}
	defer conn.Close()

	//Test setup
	client := NewSimpleServiceClient(conn)

	//Test transactions
	start := time.Now()
	for {
		name = <-ch
		printMsg("say hello to: " + name)

		err := testTrans(appCtx, client)

		if err != nil {
			log.Fatalf("Test transaction ailed; Details: %s", err)
		}
		fmt.Printf("SUCCESS running transactions! (average %s per operation)\n", time.Duration(float64(time.Since(start))))
	}

	select {}
}

var name string = "Websocket Test Client"

func printMsg(s string) {
	p := htDoc.Call("createElement", "p")
	p.Set("textContent", s)
	elBody.Call("appendChild", p)
}

func testTrans(ctx context.Context, client SimpleServiceClient) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	reaper, err := client.ReapServerStream(ctx, &types.Empty{})
	if err != nil {
		return fmt.Errorf("Could not get server stream; Details: %w", err)
	}

	for {
		reply, err := reaper.Recv()
		if err != nil {
			if err != io.EOF {
				printMsg(err.Error())
				return fmt.Errorf("Could not read stream; Details: %w", err)
			}
			break
		}
		printMsg(reply.Message)
	}
	return nil
}
