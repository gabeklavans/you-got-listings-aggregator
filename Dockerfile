FROM python:3.13-alpine

RUN apk add --no-cache go python3 py3-pip libxml2 libxslt

WORKDIR /app

COPY . .

RUN pip3 install -r ./scraper/requirements.txt

RUN go build -o main .

ENTRYPOINT ["./main"]

