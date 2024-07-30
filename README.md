# Go-Backup

Go-Backup is a simple command line utility written in Go that leverages [rclone](https://rclone.org) to transfer files to cloud storage.

It uses a JSON configuration file in which you can define custom behavior for multiple devices: for each machine you use the same configuration with, you can specify a list of directories and/or files to send, along with a list of pre-transfer commands and/or a list of post-transfer commands to be executed locally.

To keep track of unattended executions, it can be configured to send the user status notifications and/or signals to external uptime monitoring services.

###### This is how I chose to learn Go, it's a practice project. You can achieve the same with a bash script and rclone itself.

## Install

### 📋 Requirements

- [Go](https://golang.org/doc/install) installed
- **Supported remotes**: WebDAV, Drive, Dropbox, S3-compliant services.

### 📦 Releases

Pre-built binaries are available in the [releases section](https://github.com/0x07cf-dev/Go-Backup/releases).

## Configuration

The first thing you need is a destination for your files. These destinations (remotes) are configured interactively using rclone, a versatile command-line program for managing files on cloud storage.

If you don't have any rclone remote configured, run the program:

```sh
go-backup upload
```

You'll be prompted to create one if none can be found. You will need an URL and your access credentials, such as username and password or access token.

###### For service-specific instructions see: [rclone configuration](https://rclone.org/docs/#configure).

### 📂 Paths and Commands

Configuring the program involves editing a JSON file. It expects a list of zero or more paths, and two lists of zero or more commands.

Each machine will create its own directory under the specified root on the remote destination. Within this directory, all paths specified for backup are recreated. On your chosen remote, you will have the structure `/Root/Hostname/...`

To clarify, let's consider the following configuration:

```json
{
  "machines": [
    {
      "hostname": "Debian01",
      "paths": [
        "/etc/important/"
        "/var/log/useful",
      ],
       ...
    }
  ]
}
```

Now let's assume you run the program with this configuration.

```sh
go-backup upload MyDrive -r "MyBackups"
```

- The machine running the program is: `Debian01`
- You specify the remote: `MyDrive`
- You specify the root: `MyBackups`

With these values, the directories `/etc/important` and `/var/log/useful` will be transferred from `Debian01` to your remote named `MyDrive` at:

```
(MyDrive) /MyBackups/Debian01/etc/important
(MyDrive) /MyBackups/Debian01/var/log/useful
```

This structure ensures that backups are organized by device and retain their original paths on the remote destination.

### The use of environment variables in paths and commands is supported.

```json
{
  "machines": [
    {
      "hostname": "windows01",
      "paths": [
        "C:\\Users\\Admin\\Documents\\"
        "%USERPROFILE%/Desktop/Stuff",
      ],
       ...
    },
    {
      "hostname": "debian02",
      "paths": [
        "/etc/mysql"
        "$HOME/stuff",
      ],
       ...
    },

     ...
  ]
}
```

## Environment

Go-Backup utilizes environment variables to setup notifications and health monitoring.<br />These variables can be set directly in your system's environment or within an environment file, using the appropriate flag.

### 📑 Using a File

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

- `go-backup -e="/path/to/.env" ...`
- `go-backup --envFile="/path/to/.env" ...`

### 🔔 Notification Setup

Go-Backup sends notifications via ntfy.sh. You can use the [public instance](https://ntfy.sh/app) or you can self-host your own.<br />To configure notifications, set the following environment variables:

- `NTFY_HOST`: The host of the ntfy server (e.g., "https://ntfy.example.com").
- `NTFY_TOPIC`: The topic to which notifications will be sent. (e.g., "mybackups")
- `NTFY_TOKEN` (**Optional**): The authentication token for the ntfy server, if required.

### 🩺 Health Monitoring

Additionally, you can optionally set to ping uptime monitoring services using one (or all) of the following variables:

- `NTFY_HEALTHCHECKS`: The ID of your Healthchecks.io monitoring service. (The part after `hc-ping.com/`)
- `NTFY_BETTERUPTIME`: The ID of your Better Uptime monitoring service. (The part after `api/v1/heartbeat/`)

These are the unique strings in the URLs that these services give you to ping.

#### For example, using Better Uptime's heartbeats:

- You have a Better Uptime account and you create a new heartbeat.
- You are given the URL: `https://uptime.betterstack.com/api/v1/heartbeat/abcdefghijklmnopqrstuvwxyz` <-- Notice the unique ID
- You set the env variable: `NTFY_BETTERUPTIME=abcdefghijklmnopqrstuvwxyz`
- You run Go-Backup. When it finishes, Go-Backup will send a heartbeat to *Better Uptime*.

## Flags

| Category | Flag | Shorthand | Description |
|----------|------|-----------|-------------|
| Execution | \--unattended | \-U | Set this to false to skip user input. Required to run via CRON/automatically. |
|  | \--simulate | \-S | Set whether the backup session should be simulated. |
|  | \--debug |  | Enables debug mode. |
|  |  |  |  |
| File Paths | \--remoteRoot | \-r | Specify the root backup directory on the remote. |
|  | \--envFile | \-e | Path to the environment file. |
|  | \--configFile | \-c | Path to the configuration file. |
|  | \--langFile |  | Path to a custom language file. |
|  |  |  |  |
| Other | \--logFile | \-o | Path to the log file. |

## Disclaimer

Please note that this is a naive and rudimentary solution designed for my small-scale need. It lacks advanced features such as version control and comprehensive error handling.

This project is the result of my limited experience and is not intended to be a robust or feature-rich solution. As such, it may not meet the requirements of more complex backup scenarios.

Please consider using already established backup solutions for your critical data management needs. This project is provided as-is, with no guarantees of suitability or reliability.

I started this project because:

- I needed to upgrade my bash script that achieved the same, but was becoming nightmarish to maintain.
- I am learning Go, so don't expect high quality code. It is what it is. 😞

## Contributing

Contributions are welcome! Feel free to submit issues or pull requests.

#### Planned:

- Interactive configuration

---