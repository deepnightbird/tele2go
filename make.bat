set GOARCH=386
set GOOS=windows
for /f "skip=1" %%x in ('wmic os get localdatetime') do (
    set tmpDate=%%x
    rem echo s %%x !tmpDate!
    goto f
)
:f
rem echo s1 %tmpDate%

set YYYYMMDD=%tmpDate:~0,4%.%tmpDate:~4,2%.%tmpDate:~6,2%
set HHMMSS=%tmpDate:~8,2%:%tmpDate:~10,2%:%tmpDate:~12,2%
set fullstamp=%YYYYMMDD%_%HHMMSS%
echo fullstamp: "%fullstamp%"


rem set flags="-ldflags -s -w"
rem set flags="-o tele2go.exe"
if exist build\tele2go.exe ( del build\tele2go.exe )
go build -o build\tele2go.exe -ldflags "-s -w -X main.BuildTime=%fullstamp%" tele2go.go ansi.go types.go version.go
if exist build\tele2go.exe (
    "F:\Program Files (x86)\upx-3.96-win32\upx.exe" -9 -k build\tele2go.exe
)
