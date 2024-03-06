# minserver

`minserver` is a small package that wraps `http.Server` and provides some very
basic and common middleware tools, like a request logger and default request
timeout. Otherwise, it constraints `http.Server` to common operations and good
defaults.
