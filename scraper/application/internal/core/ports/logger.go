package ports

type Logger interface {
	LogDebug(requestId string, operation string)
	LogError(requestId string, err error)
	Close()
}
