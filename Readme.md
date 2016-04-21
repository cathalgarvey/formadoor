# FormaDoor
### Simple, Scaleable, Keyless Door Access for Hackerspaces and Clubhouses
by Cathal Garvey, Copyright 2016, Released under the AGPLv3 or later.

## Disclaimer
This is a piece of software designed to control access to presumably-important resources, such as a building. I, Cathal Garvey, take no responsibility for your usage of this software, and offer no warranties of any kind that it works at all or is fit for any purpose. If you use this and get robbed, locked out, or the building catches fire, that's on you, not me.

## About
[Forma Labs](http://formalabs.org) is Ireland's first Biohackerspace, based in Cork City and providing a creative space for biotechnology for curious folks.

Because we want to be able to offer easy access and invite memberships to newcomers with little friction, we want to offer door access to people quickly. However, doing this with keys means managing key cutting, tracking who has keys, and then recovering keys from lapsing or leaving members or from people who, for whatever reason, can no longer be trusted with a key (we got unlucky, this happened early in Forma Labs' lifetime).

My solution was to rig up our electronic latch to a Raspberry Pi and configure it to accept time-based one-time-pad keys, or TOTP, which can be generated on smartphones and typed into a USB keyboard fed out the doorframe.

This solution has been used for about a year in Forma Labs but until now was written in unmaintainable Python. It has recently been rewritten into a Go server for the PiFace controller and a CLI client that accepts and validates codes. This not only allows for better static checks of program correctness, it broadens scope for other methods of door validation in future, including emails, SMS, or one-click mobile Apps. It has somewhat increased the complexity of simple set-ups, though. See below for a setup guide.

### Features
#### Door Control Server
1. Logging of door access requests by API key.
2. HMAC & Timestamp secured request interface.
    * Requests must contain a JSON object with a timestamp and a requested lease.
    * This JSON body must be authenticated by HMAC placed in the request header.
    * This system isn't designed to scale, yet (nor is it likely to): All API keys are tested against the header right now.
3. Failed attempts rate-limit for several seconds.

#### TOTP & Time-period CLI Client
1. TOTP-based authentication with the door system through numeric keypad.
2. Day-of-week and time-of-day based time-framing to ensure access only in specified periods.
3. Logging of door access attempts and successful logins, by name.
4. Configuration by simple JSON file entries.
5. Forgiving TOTP lease time allows for the use of just-prior keys, preventing the "wait for next key" antipattern when the TOTP pie-chart is nearly finished.

### Usage
1. Configure your Raspberry Pi and Piface, or equivalent system (the door server needs a rewrite to accept a door-control interface to broaden scope from PiFace..)
2. Configure your Pi to auto-login as the "pi" user. This means `.bashrc` will be executed, so the rest of the configuration can be performed there.
3. Create a folder named `doorcontrol` in your home folder for user "pi", and place the following there:
    * `apiTokens.json` - A list of JSON objects containing API token information for the door service. At least one is necessary for the CLI client. Each object must have `Key`, `Name`, `DevName`, `DevEmail` keys, all strings. Key can be anything; it's used as a HMAC secret so make it at least 32 properly random bytes for security.    
    * `cliToken.txt` - A file containing only the CLI API token/key from above, with no newline.
    * `cliAuthSecrets.json` - A list of JSON objects containing CLI TOTP authentication secrets and user details. Each object consists of string keys `name`, `time policy`, `secret`, `email`. Time policy is of form "[Dow:Dow]HH:MM->HH:MM" or optionally a bar-separated list of such policies, such as `[Sat:Sun]12:00->17:00|[Mon:Fri]08:45->18:30`. Secret is the TOTP secret, encoded in uppercase base32.
4. Add two lines to your `.bashrc` to start the server and the CLI client, and capture logging output:
    * `doorMicroservice $HOME/doorcontrol/apiTokens.json >> $HOME/doorLogs.txt &`
    * `totpClient $HOME/doorcontrol/cliAuthSecrets.json "$(cat $HOME/doorcontrol/cliToken.txt)" >> $HOME/doorLogs.txt`
5. Build `doorMicroservice` and `totpClient` (from clitools directory) for the Raspberry Pi and copy them to `/usr/bin` on the door controller Pi.
6. Restart or Ctrl-D to kick off the new `.bashrc` and launch the two services.
7. Provision your members with QR codes for the TOTP tokens as usual and instruct them to use secure, open source tools to calculate tokens like the older open version of Google Authenticator or some similar tool from the [F-Droid open source Android store](https://f-droid.org).
8. Ensure numlock is enabled on that USB keypad you tacked to the wall outside! I have plans to push code that will interpret the non-numlock output as numbers for the CLI client but right now Numlock is a leading cause of n00b phonecalls from members..
