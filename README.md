# Go-Backup -

Go-Backup is a simple command line utility written in Go that leverages [rclone](https://rclone.org) to transfer files to a remote destination.

It uses a JSON configuration file to customize behavior for each device you run it on. For each of your machines, you can specify a list of directories and/or files to send, along with lists of pre-transfer and/or post-transfer commands to be executed. It can be configured to send the user status notifications and/or signals to external uptime monitoring services, to keep track of non-interactive executions.

###### This is how I chose to learn Go. You can achieve the same with a script and rclone itself.

## Install
### Requirements
- [Go](https://golang.org/doc/install) installed
- **Supported backends**: WebDAV, Drive, Dropbox, S3-compliant services.

### Releases
Pre-built binaries for Go-Backup are available in the [Releases](https://github.com/0x07cf-dev/Go-Backup/releases) section of this repository.

### Build from Source
To build Go-Backup from source, follow these steps:

1. **Clone the Repository:**
   Clone the Go-Backup repository to your local machine:

   ```sh
   git clone https://github.com/0x07cf-dev/Go-Backup.git
   ```

2. **Navigate to the Directory:**
   Change your current directory to the cloned repository:

   ```sh
   cd Go-Backup
   ```

3. **Build the Executable:**
   Build the Go-Backup executable using the following command:

   ```sh
   go build -o ./go-backup .
   ```

   This will generate the `go-backup` executable in the current directory.

4. **Run:**
   Once built, you can run Go-Backup using the following command:

   ```sh
   ./go-backup
   ```

   Optionally, you can specify flags to customize the execution, such as:

   ```sh
   ./go-backup --simulate --unattended upload MyRemote
   ./go-backup -S -U upload Drive
   ```

   For a list of flags, refer to the Flag section below, or run ```./go-backup --help```

## Configuration

Configuring Go-Backup is straightforward. You can specify 

### 1.  Generate Default Configuration:  
Run the program for the first time to generate a default configuration file. This file will serve as the template for defining backup behavior for each device.  

### 2. Edit the Configuration File:  
You can now edit the configuration file. Provide the appropriate paths and commands for pre and post-transfer operations.

You can use the same configuration file for as many devices as you'd like. *Do not modify the hostname!*

#### Using environment variables in your paths is supported depending on the OS:

\$MY_VAR on Linux, %MY_VAR% on Windows

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

Go-Backup utilizes environment variables to setup notifications and health monitoring.<br>These variables can be set directly in your system's environment or within an environment file, using the appropriate flag.

### 📑 Using Environment Files
You can declare environment variables in an `.env` file for convenience. They are defined in a `KEY=value` format.

Here's an example `.env` file:

```plaintext
NTFY_HOST=https://ntfy.example.com
NTFY_TOKEN=tk_abcdefghijklmnopqrstuvwxyz
NTFY_TOPIC=tests

NTFY_HEALTHCHECKS=aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee
NTFY_BETTERUPTIME=abcdefghijklmnopqrstuvwxyz
```

You can specify the path to an  environment file by running the program with the flags: 
- `-e=/path/to/.env`.
- `--envFile=/path/to/.env`


### 🔔 Notification Setup
Go-Backup sends notifications via ntfy.sh. You can use the [public server](https://ntfy.sh/app) or you can self-host your own instance. Both options work.<br>To configure notifications, set the following environment variables:

- `NTFY_HOST`: The host of the ntfy server (e.g., "https://ntfy.example.com").
- `NTFY_TOPIC`: The topic to which notifications will be sent. (e.g., "mybackups")
- `NTFY_TOKEN` (**Optional**): The authentication token for the ntfy server, if required.

### 🩺 Health Monitoring
Additionally, you can set up some optional uptime monitoring services using one (or all) of the following variables:

- `NTFY_HEALTHCHECKS`: The ID of your Healthchecks.io monitoring service. (The part after `hc-ping.com/`)
- `NTFY_BETTERUPTIME`: The ID of your Better Uptime monitoring service. (The part after `api/v1/heartbeat/`)

These are the unique strings in the URLs that these services give you to ping.

#### For example, using Better Uptime's heartbeats:
- You have a Better Uptime account and you create a new heartbeat.
- You are given the URL: `https://uptime.betterstack.com/api/v1/heartbeat/`***`abcdefghijklmnopqrstuvwxyz`*** <-- Notice the unique ID
- You set the env variable: `NTFY_BETTERUPTIME=abcdefghijklmnopqrstuvwxyz`
- You run Go-Backup. When it finishes, Go-Backup will send a heartbeat to *Better Uptime*.


## Flags

| Category   | Flag            | Shorthand | Description |
|------------|-----------------|-----------|-------------|
| Execution  | --unattended   | -U        | Set this to false to skip user input. Required to run via CRON/automatically. |
|            | --simulate     | -S        | Set whether the backup session should be simulated. |
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
