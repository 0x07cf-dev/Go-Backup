# Go-Backup

Go-Backup is a simple utility written in Go that uses rclone to transfer files to a remote destination.

It follows a .json configuration in which you can define custom behaviour for each device you run it on. You can specify which directories and/or files to transfer, along with pre and/or post-transfer commands to be executed on the machine.

It can optionally be configured to send status notifications to the user via [ntfy.sh](https://ntfy.sh/app), and/or heartbeat signals to external uptime monitoring services in order to keep track of non-interactive executions.
###### This is how I chose to learn Go. 

## Configuration

To configure Go-Backup, you first need to run it so that the default configuration can be generated:

```json
{
    "machines": [
        {
            "hostname": "automatically-generated",
            "paths": [],
            "output": true, 
            "pre": [],
            "post": []
        }
    ]
}
```

Now you can edit the config file. *Do not modify the hostname!*

Ensure you provide the appropriate paths and commands for pre and post-transfer operations.
#### You can also use environment variables in your paths:

```json
{
    "machines": [
        {
            "hostname": "your-windows-machine",
            "paths": [
                "%USERPROFILE%\\Desktop\\Stuff",
                "C:\\Users\\Admin\\Documents\\"
            ],
            "output": true, 
            "pre": [],
            "post": []
        },
        {
            "hostname": "your-linux-machine",
            "paths": [
                "$HOME/stuff",
                "/etc/stuff/some.conf"
            ],
            "output": true, 
            "pre": [
                "mariadb-dump --databases db1 > /path/to/output.sql"
            ],
            "post": [
                "rm /path/to/remove/*"
            ]
        }
    ]
}
```

## Environment

Go-Backup utilizes environment variables for notifications and health monitoring.<br>These variables can be set directly in your system's environment or within an `.env` file for convenience.

### 🔔 Notification Setup

Go-Backup sends notifications via ntfy.sh. You can use the [public server](https://ntfy.sh/app) or you can self-host your own instance. Both options work.<br>To configure notifications, set the following environment variables:

- `NTFY_HOST`: The URL of the ntfy server (e.g., "https://ntfy.example.com").
- `NTFY_TOPIC`: The topic to which notifications will be sent.
- `NTFY_TOKEN`: The authentication token for the ntfy server.

### 🩺 Health Monitoring

Additionally, you can set up some optional uptime monitoring services using one (or all) of the following variables:

- `HEALTHCHECKS_ID`: The ID of your Healthchecks.io monitoring service. (The part after `hc-ping.com/`)
- `BETTERUPTIME_ID`: The ID of your Better Uptime monitoring service. (The part after `api/v1/heartbeat/`)

These IDs are the unique strings in the URLs these services give you.

##### For example, using Better Uptime heartbeats:
- You are given this URL: `https://uptime.betterstack.com/api/v1/heartbeat/abcdefghijklmnopqrstuvwxyz`
- Your ID will be: `abcdefghijklmnopqrstuvwxyz`.
- You will need to set: `BETTERUPTIME_ID="abcdefghijklmnopqrstuvwxyz"`
- The program will send a heartbeat to Better Uptime at the end of the session.

### 📑 Using an Environment File (.env)

You can also store these environment variables in an `.env` file for convenience. You can specify the path to an  environment file by running the program with the flag: `--envFile=/path/to/.env`.

Here's an example `.env` file:

```plaintext
NTFY_HOST=https://ntfy.example.com
NTFY_TOKEN=tk_abcdefghijklmnopqrstuvwxyz
NTFY_TOPIC=tests
HEALTHCHECKS_ID=aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee
BETTERUPTIME_ID=abcdefghijklmnopqrstuvwxyz
```

## Flags

| Category   | Flag          | Description |
|------------|---------------|-------------|
| Execution  | --interactive | Set this to false to skip user input.<br>**Required to run via CRON/automatically**. |
|            | --simulate    | Set whether the backup session should be simulated. |
|            | --debug       | Enables debug mode. |
| File Paths | --remoteRoot  | Specify the root backup directory on the remote. |
|            | --envFile     | Specify the path to the environment file. |
|            | --configFile  | Specify the path to the configuration file. |
|            | --langFile    | Specify the path to a custom language file. |
|            | --logFile     | Specify the path to the log file. |

## Disclaimer

It's important to note that this is a naive and rudimentary solution designed for small-scale needs. It lacks advanced features such as version control and comprehensive error handling.

This project is the result of my limited experience and is not intended to be a robust or feature-rich solution. As such, it may not meet the requirements of more complex backup scenarios.

Please consider using established backup solutions for critical data management needs. This project is provided as-is, with no guarantees of suitability or reliability.

I started this project because:
- I needed to upgrade my bash script that achieved the same, but was becoming nightmarish to maintain.
- I am learning Go, so don't expect high quality code. It is what it is. 😞


## Contributing

Contributions are welcome! Feel free to submit issues or pull requests.


#### To-do:
- Use cobra+viper
- Better path/command parsing
- Log rotation?
- Other to-do's

---