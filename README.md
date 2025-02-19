# GitHub Notifier 🫎

## Introduction
When working with GitHub, you most likely have spent several minutes sometimes hours waiting for the GitHub workflow to finish or waiting for a Pull request to be reviewed. This project aims to minimize that gaps of waiting for external.


## Features
- 👥 Receive a notification when my teammate request a review in a PR
- 📝 Receive a notification when someone reviews my PR
- 🧾 Receive a notification when the checks of a PR fail/succeed

## How it works
### Notifications
On startup, we check the unreviewed PRs of the last 24 hours. After that, every 1 minute, we check any unreviewed PR created in the last minute.
### Reviews and Checks
On startup, we check the new reviews and checks that were completed in the last 24 hours. After that, every 1 minute, we check any new reviews or checks in the last minute.

## Getting started

### Prerequisites
- Go 1.22.x or later
- Git

### Installation
1. Install Go from the [official website](https://golang.org/dl/).
2. Install the binary using `go install`:
   ```sh
   go install github.com/wassimbenzarti/github-notifier@latest
   ```

### Configuration
1. Set up the environment variables:
   ```sh
   export GITHUB_TOKEN=your_github_token
   ```
2. Run the `init` command to create the configuration file:
   ```sh
   github-notifier config init
   ```

### Running the application
1. Start the application:
   ```sh
   github-notifier run
   ```

# Known limitations
1. One person in the team reviewing all PRs might eventually impact the quality of code (this tool doesn't notify you about already reviewed PRs)
2. Notifications can always be a subjectively distracting, so you have to use it wisely. Currently this tool doesn't support features to limit the distractions (e.g. only alert if there are more than 3 events)
3. If you review a PR but don't approve it, the owner of the PR needs to request your review again (by clicking the "request again" button on GitHub) to get a notification
