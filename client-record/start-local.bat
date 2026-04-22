@echo off
setlocal

set HOST=https://h5.chichengwld.com/
set DY_HOST=https://h5.chichengwld.com/
set IMG_HOST=https://h5.chichengwld.com/
set ROUTER_BASE=/
set PORT=8081
set WEB_TOKEN=eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiIsImp0aSI6IjRmMWcyM2ExMmFhYmxpenppIn0.eyJpc3MiOiJodHRwczpcL1wvaDUuY3ltc3kuY25cLyIsImF1ZCI6Imh0dHBzOlwvXC9oNS5jeW1zeS5jblwvIiwianRpIjoiNGYxZzIzYTEyYWFibGl6emkiLCJpYXQiOjE3NzY3NjA2ODAsIm5iZiI6MTc3Njc2MDY4MCwiZXhwIjoxNzc3MDE5ODgwLCJ1c2VyX2luZm8iOiJ7XCJ1c2VyX2lkXCI6NDc3NTM5MyxcInVzZXJfbmFtZVwiOlwidWF0X3Rlc3Q0RkdEWjRBODAxMDM3MjVcIixcImdhbWVfaWRcIjpcIjcwMThcIixcImFnZW50X2lkXCI6Nzc0NyxcInRvcF9hZ2VudF91aWRcIjoxNTkzLFwiY3Vycm5lY3lpZFwiOlwiLVwifSJ9.7tQ0NOtndkKPzrDyENhCK4fz68XvA6yfQ6XlVztKL0Y
set DEBUG_URL=http://127.0.0.1:%PORT%/?token=%WEB_TOKEN%&language=zh-cn&timezone=8

echo Dev server URL:
echo %DEBUG_URL%
echo.
echo Login.vue will read token/language/timezone from the query string.
echo Keep this window open while webpack-dev-server is running.
echo.

call npm run dev
