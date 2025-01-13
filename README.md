# YGL Aggregator

Go server + website + Python scraper designed to manage YGL listing links received from brokers and notify when new listings are added.

## Run with Docker

- Populate your `.env` file (take a look at `.env-example`)

- For now, the only option is to build from source
```sh
docker compose up -d
```

## Develop

### Setup

- Setup your *.env* file

```sh
cp .env-example .env
```
then fill your own values into *.env*

- Install system deps

Debian-based systems (Ubuntu)
```sh
sudo apt-get build-dep python3-lxml
```
RHEL-based systems (Fedora)
```sh
sudo dnf in python3-devel libxml2-devel libxslt-devel
```

- Install python dependencies (recommended to use a virtual environment)
```sh
python -m venv .venv
source ./.venv/bin/activate # run `deactivate` to turn off the venv, or just `exit` the terminal
pip install -r ./scraper/requirements.txt
```

### Run

- Start the server
```sh
go run .
```

- Run the scraper manually (in a seperate terminal, with your venv activated if used)
```sh
./scraper/main.py --db ./ygl.db --sites ./sites-example.json
```

