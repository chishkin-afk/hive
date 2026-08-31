# Custom logger
*based on slog*

## Install
```
git submodule add https://github.com/braunkc/log
```

## Quick start
``` go
func main() {
	cfg := &log.Config{
		Service:     "service_name",    // optional
		OutputType:  log.Both,          // required
		LogFilePath: "./file_name.log", // optional, default: root of project
		Level:       slog.LevelDebug,   // optional, default: slog.LevelInfo
		JSONFormat:  false,             // optional, default: false
	}

	handler, err := log.NewHandler(cfg)
	if err != nil {
		panic(err)
	}
	l := slog.New(handler)

	l.Debug("debug message", slog.Any("key", "value"))
	l.Info("info message", slog.Any("key", "value"))
	l.Warn("warn message", slog.Any("key", "value"))
	l.Error("error message", slog.Any("key", "value"))
}
```

## Config
``` go
type OutputType int

const (
	Console OutputType = 1
	File    OutputType = 2
	Both    OutputType = 3
)

type Config struct {
	Service     string     // service title
	OutputType  OutputType // console/file/both
	LogFilePath string     // log file path
	Level       slog.Level
	JSONFormat  bool       // console output format
}
```

## Features
* colorful output to console with emoji
* flexible config
* timestamp in UTC

## Examples
### Console output
```
14:00:41.000 🐛 DEBUG debug message key:value service:service_name
14:00:41.000 ℹ️ INFO info message key:value service:service_name 
14:00:41.000 ⚠️ WARN warn message key:value service:service_name 
14:00:41.000 ❌ ERROR error message key:value service:service_name
```

### JSON format
``` json
{"time":"2025-05-31T14:00:41.000000Z","level":"DEBUG","source":{"function":"func_name","file":"C:path_to_file.go","line":41},"msg":"debug message","key":"value"}
{"time":"2025-05-31T14:00:41.000000Z","level":"INFO","source":{"function":"func_name","file":"C:path_to_file.go","line":41},"msg":"info message","key":"value"}
{"time":"2025-05-31T14:00:41.000000Z","level":"WARN","source":{"function":"func_name","file":"C:path_to_file.go","line":41},"msg":"warn message","key":"value"}
{"time":"2025-05-31T14:00:41.000000Z","level":"ERROR","source":{"function":"func_name","file":"C:path_to_file.go","line":41},"msg":"error message","key":"value"}

```