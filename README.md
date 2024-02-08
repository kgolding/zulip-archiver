# Zulip archiver to SQLite database

Super simple way to archive your Zulip chat messages and files.

## Installation

1. Install Golang
2. `git clone git@github.com:kgolding/zulip-archiver.git`
3. `cd zulip-archiver`

## Usage

1. Create a Zulip Bot with an email address and make a note of the API key.
2. `go run main.go data <host> <email> [<api_key>]` to download the messages
3. `go run main.go files <host> <email> [<api_key>]` to download the avatars and files referenced to from within the messages (limited to files on the same host)

    * `host`: zulip server domain/IP e.g. chat.myzulipserver.com
    * `email`: address of the user account
    * `api_key`: API key for the given email

The download takes some time, as there are limits on the Zulip APIs, so we pause every second between calls which happen once for each topic or 200 messages thereof.

## Work in progress

* [X] Connect to Zulip server and download streams, topics & messages
* [X] Download streams and save to DB
* [X] Download messages and save to DB
* [X] Download avatar images
* [X] Download images/files referenced to in messages
* [ ] Create web portal to view streams, messages, files
* [ ] Implement search feature in web portal

