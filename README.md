<div align = "center">
  <img alt = "project logo" src = "logo.png" />
  <h1>Faridoon</h1>

  Easily save and publish your favourite chat quotes for others to see.

[![Static Badge](https://img.shields.io/badge/maturity-Production-brightgreen)](#none)
![GitHub Actions Workflow Status](https://img.shields.io/github/actions/workflow/status/jamesread/Faridoon/release-pipeline.yml?link=https%3A%2F%2Fgithub.com%2Fjamesread%2Ffaridoon%2Factions%2Fworkflows%2Frelease-pipeline.yml)
[![Discord](https://img.shields.io/discord/846737624960860180?label=Discord%20Server)](https://discord.gg/jhYWWpNJ3v)

Documentation for how to install and use Faridoon is available at https://jamesread.github.io/Faridoon/

</div>

## Screenshots

### Latest quotes

![Latest quotes](var/marketing/screenshots/quotes.png)

The front page lists the newest quotes, with each chat username highlighted in its own colour. Signed-in users can vote quotes up or down.

### Quote permalink

![Quote permalink](var/marketing/screenshots/quote.png)

Every quote has its own page, so a favourite exchange is easy to link to and share.

### Add a quote

![Add a quote](var/marketing/screenshots/add-quote.png)

Paste a chat log and a live preview shows exactly how the quote will look before it is submitted.

### Approval queue

![Approval queue](var/marketing/screenshots/approvals.png)

Quotes from guests and regular users wait in an approval queue, where moderators can approve, reject or tidy them up before they go live.

### Settings

![Settings](var/marketing/screenshots/settings.png)

Admins configure the site from the browser, including the site title, voting, registration, guest submissions and theme, with no config files to edit.

### Dark mode

![Dark mode](var/marketing/screenshots/quotes-dark.png)

A built-in dark theme, which users can switch to from the header.

Screenshots use fictional test data and are regenerated with `make screenshots`
(see [`var/marketing/screenshots`](var/marketing/screenshots/README.md)).

## Getting Started

Documentation is available at: https://jamesread.github.io/Faridoon/ , the recommended installation method is via **Docker**, connected to a **MySQL Database**. Faridoon runs just fine with low system resources - 1 vCPU and a 1 GB of RAM is plenty.

## Stack

* Go service with ConnectRPC
* Vue 3 + Vue Router + [picocrank](https://www.npmjs.com/package/picocrank) UI components
* MySQL schema managed with **sql-migrate**
* Authentication via [httpauthshim](https://github.com/jamesread/httpauthshim) (sessions; users stored in MySQL)

## Features

* Automatic highlighting of usernames.
* Semi-intelligent removal of line breaks and weird characters.
* User login and registration, with admins and privilege-based moderators.
* Approval queue for guest and non-privileged submissions (reject deletes the quote).
* Admin Settings (database cvars) for site title, voting, registration, guest add, and syntax highlighting.
* Header links, webhooks (`approval.requested`), audit logs, and diagnostics.
* YAML config for listen/database/auth; environment variables for database connectivity (and a few bootstrap values).

## **Faridoon is a No-Nonsense Open Source project;**

- All code and assets are Open Source (AGPL).
- No company is paying for development, there is no paid-for support from the developers.
- No separate core and premium version, no plus/pro version or paid-for extra features.
- No SaaS service or "special cloud version".
- No "anonymous data collection", usage tracking, user tracking, telemetry or email address collection.
- No requests for reviews in any "app store" or feedback surveys.
- No prompts to "upgrade to the latest version".
- No internet-connection required for any functionality.

## Docs

Documentation for how to install and use Faridoon is available at https://jamesread.github.io/Faridoon/
