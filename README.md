# Umbrella API

A modern Go API that tells you if you need an umbrella today based on weather data from the Australian Bureau of Meteorology.

## Features

- Fetches real-time precipitation data from BOM FTP server
- REST API endpoint returning JSON
- Clean HTML web interface
- Structured logging with zerolog
- Uses Go standard library for HTTP and FTP

## Installation

```bash
go mod download
```

## Usage

Run the server:

```bash
go run main.go
```

The server starts on port 8080 by default (configurable via `PORT` environment variable).

## Endpoints

- `GET /` - HTML interface showing umbrella recommendation
- `GET /api/umbrella` - JSON API endpoint

### JSON Response Example

```json
{
  "need_umbrella": true,
  "precipitation_chance_percent": 15,
  "location": "NSW_PT131",
  "timestamp": "2025-10-06T10:30:00+10:00"
}
```

## Logic

The API checks the precipitation probability for area NSW_PT131. If the chance is greater than 5%, it recommends taking an umbrella.

## Data Source

Weather data from: `ftp://ftp.bom.gov.au/anon/gen/fwo/IDN11060.xml`
