# loqu
A small web server/client utility that has a lot to say

# Usage
```
loqu is a web server (and client) with one simple job: log all the things!

Usage:
  loqu [command]

Available Commands:
  call        Execute calls against a web server
  help        Help about any command
  serve       Starts an HTTP server which logs all lifecycle events

Flags:
      --alsologtostderr                  log to standard error as well as files
      --config string                    config file (default is $HOME/.loqu.yaml)
  -h, --help                             help for loqu
      --log_backtrace_at traceLocation   when logging hits line file:N, emit a stack trace (default :0)
      --log_dir string                   If non-empty, write log files in this directory
      --logtostderr                      log to standard error instead of files (default true)
      --stderrthreshold severity         logs at or above this threshold go to stderr (default 2)
  -t, --toggle                           Help message for toggle
  -v, --v Level                          log level for V logs
      --vmodule moduleSpec               comma-separated list of pattern=N settings for file-filtered logging

Use "loqu [command] --help" for more information about a command.
```
