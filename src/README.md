# SRC Directory

## Air

Contains the data structures and utilities for retrieving forecasts from the AirNow API.

## dialer

Sends every API request, with a timeout, and rewrites network failures for people (naming the host, never the URL, which can carry an API key).

## geolocation

Finds coordinates from an IP address (ipwho.is) or a US zip code (api.zippopotam.us), and validates zip codes.

## report

Formats data from API calls into specific reports for display in terminal.

## storage

Reads and writes the config file and the forecast cache, in the XDG config and cache directories. Every write is atomic.

## weather

Contains the data structures and utilities for retrieving weather forecasts from the NOAA National Weather Service API.
