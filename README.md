# go-mine-stats
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
        "max_response_limit": 100,
        "default_world": "world",
        "port": ":3000"
    },
    "polling": {
        "polling_speed": 5,
        "invertBlacklist": false,
        "blacklist": {
            "operators": false,
            "banned": false,
            "list": []
        },
        "fetch_mojang_usernames": true
    }
}
```
* `server_list`:
  * `server_path`: The complete file path to where your Minecraft server directory is located.
  * `world_name`: The name of the Minecraft world whose statistics you want to scan.
* `api`:
  * `default_response_limit`: The default amount of statistics entries the API will return if `limit` is not specified in the query.
  * `max_response_limit`: The maximum amount of statistics entries the API will return.
  * `default_world`: The default Minecraft world that the API should return if `world` is not specified in the query. You may leave this blank if you don't want this behavior.
  * `port`: The port the API will listen on.

* `polling`:
  * `polling_speed`: How often (in minutes) the program should check player statistics files for changes. Recommend keeping this relatively high, as the database will store every change made to a statistic as a new row.
  * `invert_blacklist`: Turn the blacklist into a whitelist.
  * `blacklist`: UUIDs whose statistics stop being scanned into the database. **Note:** This will NOT remove player data already scanned inside the database.
    * `operators`: Stop logging stats from operators.
    * `banned`: Stop logging stats from banned players.
    * `list`: Arbitrary list of  blacklisted UUIDs
  * `fetch_mojang_usernames`: Allow go-mine-stat to fetch players' usernames from Mojang's API if not found in local files.

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