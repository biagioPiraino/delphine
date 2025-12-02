package main

func main() {
	//logger.InitLogger()
	crawlerEngine := NewApp(Config{BrokerAddress: "localhost:9094"})
	crawlerEngine.Run()
}
