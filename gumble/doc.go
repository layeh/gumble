// Getting started
//
//1. Create a new `gumble.Config` to hold your connection settings:
//
//        config := gumble.Config{
//          Username: "gumble-test",
//          Address:  "example.com:64738",
//        }
//
//2. Create a new `gumble.Client`:
//
//        client := gumble.NewClient(&config)
//
//3. Implement `gumble.EventListener` (or use gumbleutil.Listener) and attach it to the client:
//
//        client.Attach(gumbleutil.Listener{
//          TextMessage: func(e *gumble.TextMessageEvent) {
//            fmt.Printf("Received text message: %s\n", e.Message)
//          },
//        })
//
//4. Connect to the server:
//
//        if err := client.Connect(); err != nil {
//          panic(err)
//        }
package gumble
