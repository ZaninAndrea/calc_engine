env GOOS=windows go build -o build/calc-notebook_windows.exe
env GOOS=darwin go build -o build/calc-notebook_macos
env GOOS=linux go build -o build/calc-notebook_linux