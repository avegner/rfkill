# rfkill
rfkill abstraction lib for Go

# API
API is pretty simple and self-explanatory. There're no New and Close funcs. Only recurrent (thread-safe) funcs exist.

# Tests & Benches
To run all tests:
```shell
go test
```

To run all benches:
```shell
go test -run XXX -bench .
```

Env. parameters:
* `ID` - device ID, uint
* `TYPE` - device type, uint

Be careful with benches since some radio devices may get unstable after lots of state changes.
