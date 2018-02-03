@echo off

set ProjectPath=%GOPATH%\src\github.com\liuhan907\waka
set SDKPath=%ProjectPath%\sdk\WakaSDK
set ProtoName=waka

protoc %ProtoName%.proto --go_out=.
protoc %ProtoName%.proto --msg_out=%ProtoName%.pb.meta.go:.
protoc %ProtoName%.proto --csharp_out .
protoc %ProtoName%.proto --cellnet_out=.

move /Y %ProtoName%.cs %SDKPath%\WakaSDK\Waka
move /Y %ProtoName%MetaProvider.cs %SDKPath%\WakaSDK\Waka