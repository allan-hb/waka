@echo off

set ProjectPath=%GOPATH%\src\github.com\liuhan907\waka
set SDKPath=%ProjectPath%\sdk\WakaSDK
set ProtoName=four

protoc %ProtoName%.proto --go_out=.
protoc %ProtoName%.proto --msg_out=%ProtoName%.pb.meta.go:.

protoc.exe %ProtoName%.proto --csharp_out .
protoc.exe %ProtoName%.proto --cellnet_out=.
protoc.exe %ProtoName%.proto --waka_out=.

if not exist %SDKPath%\Release mkdir %SDKPath%\Release

del %SDKPath%\Release\*.proto /Q

copy /Y %ProtoName%.proto %SDKPath%\Release

if not exist  %SDKPath%\WakaSDK\Generated mkdir %SDKPath%\WakaSDK\Generated

move /Y %ProtoName%.cs %SDKPath%\WakaSDK\Generated\Generated.cs
move /Y %ProtoName%MetaProvider.cs %SDKPath%\WakaSDK\Generated\GeneratedMetaProvider.cs
move /Y CoreSupervisor.cs %SDKPath%\WakaSDK\
move /Y IDispatcher.cs %SDKPath%\WakaSDK\Generated\
move /Y Supervisor.cs %SDKPath%\WakaSDK\Generated\