@echo off

REM Copyright (c) 2020 Juan Medina.
REM
REM  Permission is hereby granted, free of charge, to any person obtaining a copy
REM  of this software and associated documentation files (the "Software"), to deal
REM  in the Software without restriction, including without limitation the rights
REM  to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
REM  copies of the Software, and to permit persons to whom the Software is
REM  furnished to do so, subject to the following conditions:
REM
REM  The above copyright notice and this permission notice shall be included in
REM  all copies or substantial portions of the Software.
REM
REM  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
REM  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
REM  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
REM  AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
REM  LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
REM  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
REM  THE SOFTWARE.
REM


IF EXIST "mesh2prod-win64.zip" (
  del "mesh2prod-win64.zip"
)

IF EXIST "mesh2prod.exe" (
  del "mesh2prod.exe"
)

cd windows

IF EXIST "mesh2prod.syso" (
  del "mesh2prod.syso"
)

windres -i mesh2prod.rc -O coff -o mesh2prod.syso

move "mesh2prod.syso" ../..

cd ..

IF EXIST "out" (
	rmdir /S /Q "out"
)

mkdir "out"

cd ..
go build -ldflags -H=windowsgui  -o build/out ./...

IF EXIST "mesh2prod.syso" (
  del "mesh2prod.syso"
)

cd build

robocopy ..\resources out\resources\ /E

tar.exe -C out -a -c -f mesh2prod-win64.zip *
