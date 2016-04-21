# Door Microservice
by Cathal Garvey, Copyright 2016, Released under AGPLv3 or later

This is a simple localhost-only microservice for opening a PiFace-powered door
latch. It listens for POST requests with a hmac header authenticating the body,
and ignores any that cannot be authenticated. Bodies are JSON objects containing
a timestamp and a request for an integer value of seconds to open the door.
Timestamps older than 5 seconds are considered invalid, and a maximum timespan
is permitted to keep the door open, by default ~10 seconds.

This is not designed to scale to many authenticated clients. HMACs right now
are verified by trying valid keys in rotation. Overlapping calls may lead to
silly things; one call's timeout callback may lock the door during another
call's door-opening lease. In theory the door should always end up locked.

Suggested implementation is to have authenticated clients be more scaleable
middlewares that serve actual members' preferred ways to access the building.
For example, the default client is a CLI utility that listens for TOTP codes
that are issued to members and carry time-based access policies per-member.
Another client might be an email listener that receives authenticated emails
from hackerspace members and permits access upon authentication in a similar
manner (authentication means timestamped encryption, not just checking
sender email header!).

## TODO
* API call JSON should have a key where clients can provide logging data
  such as the name of the authenticated member.
