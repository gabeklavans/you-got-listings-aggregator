FROM python:3.13-alpine

RUN apk add --no-cache go python3 py3-pip libxml2 libxslt

WORKDIR /app

# deps
COPY ./scraper/requirements.txt ./requirements.txt
RUN pip3 install -r ./requirements.txt

COPY ./go.mod ./go.sum .
RUN go mod download

# main app
COPY . .

RUN go build -o main .

ENTRYPOINT ["./main"]
