# Zulip archiver to SQLite database

Super simple way to archive your Zulip messages.

## Installation

1. Install Golang
2. `git clone git@github.com:kgolding/zulip-archiver.git`
3. `cd zulip-archiver`

## Usage

1. Create a Zulip Bot with an email address and make a note of the API key.
2. `go run main.go <host> <email> [<api_key>]`

* host: zulip server domain/IP e.g. chat.myzulipserver.com
* email: address of the user account
* api_key: API key for the given email or can be set in the `API_KEY` environment variable

The download takes some time, as there are limits on the Zulip APIs, so we pause every second between calls which happen once for each topic or 200 messages thereof.

## Work in progress

* [X] Connect to Zulip server and download streams, topics & messages
* [X] Download streams and save to DB
* [X] Download messages and save to DB
* [ ] Download avatar images
* [ ] Download images/files referenced to in messages

