FROM golang:1.17-alpine AS build
WORKDIR /go/conway
COPY . .
RUN cp $(go env GOROOT)/misc/wasm/wasm_exec.js ./html/wasm_exec.js
RUN GOOS=js GOARCH=wasm go build -o ./html/main.wasm .

FROM nginx:1.21-alpine
COPY html/nginx /etc/nginx 
COPY --from=build /go/conway/html/index.html  /go/conway/html/main.wasm /go/conway/html/wasm_exec.js var/www/code.sknk.ws/public/ 
RUN adduser -u 82 -D -S -G www-data www-data 
EXPOSE 80
