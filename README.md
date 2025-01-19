# YGL Aggregator

Go server + website + Python scraper designed to manage YGL listing links received from brokers and notify when new listings are added.

## Run with Docker

- For now, the only option is to build from source
```sh
docker compose up -d
```

## Develop

### Setup

- Install system deps

Debian-based systems (Ubuntu)
```sh
sudo apt-get build-dep python3-lxml
```
RHEL-based systems (Fedora)
```sh
sudo dnf in python3-devel libxml2-devel libxslt-devel
```

- Install python dependencies

> [!TIP]
> It's recommended to use a virtual environment

```sh
python -m venv .venv
source ./.venv/bin/activate # run `deactivate` to turn off the venv, or just `exit` the terminal
pip install -r ./scraper/requirements.txt
```

#### OPTIONAL

- Setup your **.env** file

```sh
cp .env-example .env
```
then fill your own values into **.env**

The **.env** file allows you to adjust the domain/port of the server

### Run

1. Start the server

> [!IMPORTANT]
> This initializes the DB that the scraper needs to exist

```sh
go run .
```

> [!NOTE]
> The server will immediately run the scraper with notifications off on startup.

2. Run the scraper manually (in a seperate terminal)

> [!IMPORTANT]
> Make sure you activate your venv first if you set one up

```sh
./scraper/main.py --db ./ygl.db
```

