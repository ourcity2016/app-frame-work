package main

import (
	"app-frame-work/client"
	"app-frame-work/context"
	"app-frame-work/handler"
	"bufio"
	oscontext "context"
	"os"
	"strings"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	wg.Add(2)
	connectionManager := context.NewSessionManagerBuilder(false, &handler.MessageHandlerImpl{})
	go func() {
		defer wg.Done()
		conn := client.BuildNewConnClient()
		conn.Conn("tcp", "127.0.0.1:8848", connectionManager, oscontext.Background())
	}()
	go func() {
		defer wg.Done()
		reader := bufio.NewReader(os.Stdin)
		for {
			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(input)
			for _, s := range connectionManager.Sessions {
				s.Handler.SendMessage(input, s.SendCh)
			}
		}
	}()
	wg.Wait()
}
