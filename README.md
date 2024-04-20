# Go-Backup
###### This is how I chose to learn Go. Even I am afraid of using it.

Go-Backup is a simple utility written in Go that uses rclone to transfer files to a remote WebDAV destination.

It follows a .json configuration in which you can define custom behaviour for each device you run it on. You can specify which directories and/or files to transfer, along with pre and/or post-transfer commands to be executed on each machine.

It can optionally be configured to send status notifications to the user via [ntfy.sh](https://ntfy.sh/app), and/or heartbeat signals to external uptime monitoring services in order to keep track of non-interactive executions.

## Configuration

To configure Go-Backup, you first need to run it so that the default configuration is generated:

```json
{
  "machines": [
    {
      "hostname": "abcdefg",
      "paths": [],
      "output": true,
      "pre": [],
      "post": []
    }
  ]
}
```

Now you can edit the config file. *Do not modify the hostname!* Ensure you provide the appropriate paths and commands for pre and post-transfer operations.

You can use the same configuration file for as many devices as you'd like.

#### You can also use environment variables in your paths:

```json
{
  "machines": [
    {
      "hostname": "windows-machine",
      "paths": [
        "%USERPROFILE%\\Desktop\\Stuff",
        "C:\\Users\\Admin\\Documents\\"
      ],
      "output": true,
      "pre": [],
      "post": []
    },
    {
      "hostname": "linux-machine",
      "paths": [
        "$HOME/stuff",
        "/etc/mysql"
        "/etc/stuff/some.conf"
      ],
      "output": true,
      "pre": [
        "mariadb-dump --databases db1 > /path/to/output.sql"
      ],
      "post": ["rm /path/to/remove/*"]
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
- `NTFY_TOKEN` (**Optional**): The authentication token for the ntfy server.

### 🩺 Health Monitoring

Additionally, you can set up some optional uptime monitoring services using one (or all) of the following variables:

- `NTFY_HEALTHCHECKS`: The ID of your Healthchecks.io monitoring service. (The part after `hc-ping.com/`)
- `NTFY_BETTERUPTIME`: The ID of your Better Uptime monitoring service. (The part after `api/v1/heartbeat/`)

These IDs are the unique strings in the URLs these services give you.

#### For example, using *Better Uptime*'s heartbeats:
- You are given the URL: `https://uptime.betterstack.com/api/v1/heartbeat/`***`abcdefghijklmnopqrstuvwxyz`***

Notice the unique string that identifies you: ***`abcdefghijklmnopqrstuvwxyz`***.
- You will need to set: `NTFY_BETTERUPTIME=abcdefghijklmnopqrstuvwxyz`
- At the end of the run, Go-Backup will send a heartbeat to *Better Uptime*.

### 📑 Using an Environment File (.env)

You can declare environment variables in an `.env` file for convenience. You can specify the path to an  environment file by running the program with the flags: 
- `-e=/path/to/.env`.
- `--envFile=/path/to/.env`

Here's an example `.env` file:

```plaintext
NTFY_HOST=https://ntfy.example.com
NTFY_TOKEN=tk_abcdefghijklmnopqrstuvwxyz
NTFY_TOPIC=tests
NTFY_HEALTHCHECKS=aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee
NTFY_BETTERUPTIME=abcdefghijklmnopqrstuvwxyz
```

## Flags

| Category   | Flag            | Shorthand | Description |
|------------|-----------------|-----------|-------------|
| Execution  | --interactive  | -i        | Set this to false to skip user input. Required to run via CRON/automatically. |
|            | --simulate     | -s        | Set whether the backup session should be simulated. |
|            | --debug        |           | Enables debug mode. |
|            |                |           | |
| File Paths | --remoteRoot   | -r        | Specify the root backup directory on the remote. |
|            | --envFile      | -e        | Path to the environment file. |
|            | --configFile   | -c        | Path to the configuration file. |
|            | --langFile     |           | Path to a custom language file. |
|            |                |           | |
| Other      | --logFile      | -o        | Path to the log file. |


## Disclaimer

Please note that this is a naive and rudimentary solution designed for small-scale needs. It lacks advanced features such as version control and comprehensive error handling.

This project is the result of my limited experience and is not intended to be a robust or feature-rich solution. As such, it may not meet the requirements of more complex backup scenarios.

Please consider using established backup solutions for critical data management needs. This project is provided as-is, with no guarantees of suitability or reliability.

I started this project because:
- I needed to upgrade my bash script that achieved the same, but was becoming nightmarish to maintain.
- I am learning Go, so don't expect high quality code. It is what it is. 😞


## Contributing

Contributions are welcome! Feel free to submit issues or pull requests.


#### To-do:
- Better path/command parsing
- Log rotation?
- Other to-do's

---
