# MiSQL

> **Note:** This is a toy project and not intended for production use.

An in-memory MySQL-compatible SQL server that loads data from YAML files. Point it at a directory of YAML files and query them with any MySQL client.

## Quick Start

```bash
docker run -v /path/to/data:/data -p 3306:3306 dlawrence00/misql:latest
```

Then connect with any MySQL client:

```bash
mysql -h 127.0.0.1 -P 3306 -u root
```

## Data Format

Each subdirectory under the data directory becomes a **database**. Each `.yaml` file within becomes a **table**.

```
data/
├── company/
│   └── users.yaml
└── cv/
    └── education.yaml
```

Example `users.yaml`:

```yaml
columns:
  - id
  - name
  - email
rows:
  - id: 1
    name: Alice
    email: alice@example.com
  - id: 2
    name: Bob
    email: bob@example.com
```

Query it:

```sql
USE company;
SELECT * FROM users;
```

## Configuration

All configuration is via environment variables:

| Variable        | Default     | Description                          |
|-----------------|-------------|--------------------------------------|
| `MYSQL_DATA_DIR`  | `./data`    | Path to directory containing databases |
| `MYSQL_HOST`      | `0.0.0.0`   | Host address to bind to              |
| `MYSQL_TCP_PORT`  | `3306`      | Port to listen on                    |

```bash
docker run \
  -v /path/to/data:/data \
  -p 3306:3306 \
  -e MYSQL_TCP_PORT=3306 \
  dlawrence00/misql:latest
```

## Memory Usage

Since all data is held in memory, you can inspect the container's memory consumption with:

```bash
# Live stats for the running container
docker stats $(docker ps -q --filter ancestor=dlawrence00/misql:latest)

# Or by container name
docker stats misql
```

Example output:

```
CONTAINER ID   NAME     CPU %   MEM USAGE / LIMIT   MEM %   ...
a1b2c3d4e5f6   misql    0.01%   18.3MiB / 15.5GiB   0.12%   ...
```

To check memory at a single point in time (non-streaming):

```bash
docker stats --no-stream misql
```

## Building from Source

```bash
git clone https://github.com/daniellawrence/misql
cd misql
go build -o misql .
MYSQL_DATA_DIR=./data ./misql
```

## Docker Image

Published image: `dlawrence00/misql:latest`

The image is built on `scratch` (no base OS) with a statically compiled binary, keeping the image minimal.
