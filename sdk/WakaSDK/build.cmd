@echo off
if not "%Platform%"=="" goto build
set currentDisk=%~d0
set currentPath="%~dp0"
if not exist "%ProgramFiles(x86)%\Microsoft Visual Studio\2017\Enterprise\VC\Auxiliary\Build\vcvars64.bat" goto community
call "%ProgramFiles(x86)%\Microsoft Visual Studio\2017\Enterprise\VC\Auxiliary\Build\vcvars64.bat"
goto build
:community
if not exist "%ProgramFiles(x86)%\Microsoft Visual Studio\2017\Community\VC\Auxiliary\Build\vcvars64.bat" goto pro
call "%ProgramFiles(x86)%\Microsoft Visual Studio\2017\Community\VC\Auxiliary\Build\vcvars64.bat"
goto build
:pro
if not exist "%ProgramFiles(x86)%\Microsoft Visual Studio\2017\Professional\VC\Auxiliary\Build\vcvars64.bat" goto error
call "%ProgramFiles(x86)%\Microsoft Visual Studio\2017\Professional\VC\Auxiliary\Build\vcvars64.bat"
goto build
:error
echo "not found visual studio 2017 any edition, now exit"
goto done
:build
cd %currentPath%
%currentDisk%

msbuild WakaSDK.sln /t:WakaSDK:Rebuild /p:Configuration=Release;Platform="Any CPU"
if ERRORLEVEL 0 goto release
:failed
echo "rebuild project failed, now exit"
goto done

:release

if not exist Release mkdir Release

copy WakaSDK\bin\Release\CellnetSDK.dll Release\CellnetSDK.dll /Y
copy WakaSDK\bin\Release\CellnetSDK.pdb Release\CellnetSDK.pdb /Y
copy WakaSDK\bin\Release\CellnetSDK.xml Release\CellnetSDK.xml /Y
copy WakaSDK\bin\Release\Google.Protobuf.dll Release /Y
copy WakaSDK\bin\Release\Google.Protobuf.xml Release /Y
copy WakaSDK\bin\Release\Google.Protobuf.pdb Release /Y
copy WakaSDK\bin\Release\WakaSDK.dll Release\WakaSDK.dll /Y
copy WakaSDK\bin\Release\WakaSDK.pdb Release\WakaSDK.pdb /Y
copy WakaSDK\bin\Release\WakaSDK.xml Release\WakaSDK.xml /Y
copy WakaSDK\bin\Release\UnitySocket.dll Release\UnitySocket.dll /Y
copy WakaSDK\bin\Release\UnitySocket.pdb Release\UnitySocket.pdb /Y
copy WakaSDK\bin\Release\UnitySocket.xml Release\UnitySocket.xml /Y

rd /s /q WakaSDK\bin
rd /s /q WakaSDK\obj

set PATH=%PATH%;C:\Program Files\7-Zip
set NOWTIME="WakaSDK_%date:~0,4%.%date:~5,2%.%date:~8,2%-%time:~0,2%.%time:~3,2%.%time:~6,2%"
7z a -t7z temp.7z ".\release\*"
rename temp.7z %NOWTIME%.7z

:done
echo done
pause > nul