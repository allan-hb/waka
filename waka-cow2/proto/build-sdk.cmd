@echo off

set ProjectPath=%GOPATH%\src\github.com\liuhan907\waka
set SDKPath=%ProjectPath%\sdk\WakaSDK

call build-proto.cmd
call %SDKPath%\build.cmd