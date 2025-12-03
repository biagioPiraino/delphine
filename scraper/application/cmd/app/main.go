package main

func main() {
	crawlerEngine := NewApp(Config{BrokerAddress: "localhost:9094"})
	crawlerEngine.Run()
}
