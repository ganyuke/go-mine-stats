# go-mine-stat
A Go-based API backend to access player statistics, inspired by [MinecraftStats](https://github.com/pdinklag/MinecraftStats). In theory, compatible with any modded statistics (such as items) as well.

**This project is in a barely usable state. Expect breaking changes often.**

# Configuration
All configuration is done through the `config.json` file found in the root of the project.

```json
{
    "server_list": [
        {
            "server_path": "../../example-mc",
            "world_name": "world"
        }
    ],
    "api": {
        "default_response_limit": 5,
        "max_response_limit": 100
    },
    "polling": {
        "polling_speed": 5,
        "blacklist": []
    }
}
```
* `server_list`:
  * `server_path`: The complete file path to where your Minecraft server directory is located.
  * `world_name`: The name of the Minecraft world whose statistics you want to scan.
* `api`:
  * `default_response_limit`: The default amount of statistics entries the API will return if `limit` is not specified in the query.
  * `max_response_limit`: The maximum amount of statistics entries the API will return.
* `polling`:
  * `polling_speed`: How often (in minutes) the program should check player statistics files for changes. Recommend keeping this relatively high, as the database will store every change made to a statistic as a new row.
  * `blacklist`: A list of UUIDs whose statistics will not be scanned into the database. **Note:** This will NOT remove player data already scanned inside the database.

## Polling statistics from multiple servers
Under `server_list`, you can add as many servers or worlds as you want by adding more JSON objects.
```json
"server_list": [
    {
        "server_path": "../../example-mc",
        "world_name": "world"
    },
    {
        "server_path": "../../paper-mc",
        "world_name": "survival"
    }
]
```
**IMPORTANT:** Make sure to change the name of your Minecraft world file to a unique name, as the database stores statistics for each world using the world name. If you fail to do this, your players' statistics may get mixed together.

# API
The API can be reached at `http://localhost:3000`, which is currently not configurable yet.

## Functions
// TODO: write when I feel like it