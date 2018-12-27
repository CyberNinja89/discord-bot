# Discord Bot

A discord bot written Go with a variety of features for a discord server. This bot utilizes [discordgo](https://github.com/bwmarrin/discordgo) library, a free and open source library.

## Table of Contents

1. [Features](#features)
1. [Usages](#usages)
1. [Building Discord Bot](#build-discord-bot)
1. [Running Discord Bot](#running-discord-bot)

## Features

- Airhorns On Demand
- Overwatch Stats
- Canned Responses

## Usages

- `!airhorn` - plays an airhorn in the current voice channel
- `!addowuser` - register the current user id with an Blizzard ID
- `!owstats` - queries the current user's Overwatch competitive ranking

## Build Discord Bot

``` bash
go build -o discord-bot ./...
```

## Running Discord Bot

``` bash
./discord-bot -t <token>
```
