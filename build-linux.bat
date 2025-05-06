@set GOARCH=amd64
@set GOOS=linux
@echo Building for Linux/amd64...
@echo ----------------------------------------------------
@echo removing old executable...
@if exist "build/linux" (
@del "build/linux" /F /Q
)
@echo creating folder structure...
@mkdir "build/linux/tinyGW"
@mkdir "build/linux/tinyGW/public/logs"
@mkdir "build/linux/tinyGW/public/webroot"
@mkdir "build/linux/tinyGW/config"
@mkdir "build/linux/tinyGW/plugin"
@echo building executable...
@go build -o "build/linux/tinyGW/zsxagw" -v -ldflags="-w -s"
@echo executable built successfully!
@echo ----------
@upx.exe "build/linux/tinyGW/zsxagw"
@echo executable compressed successfully!

@xcopy public "build/linux/tinyGW/public" /E /Y
@del "build/linux/tinyGW/public/logs/energy.log"
@echo "-----------completed!-----------"