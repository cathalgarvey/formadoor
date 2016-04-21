# TOTP Set
by Cathal Garvey, Copyright 2016, Released under AGPLv3 or later

TOTP is usually used as a secondary authentication method, tied to an explicit
person's TOTP key. For Forma's door, it is instead used as a primary
authentication method, with TOTP tokens being entered anonymously and checked
against the roster of valid user tokens.

For this to be performant, checks need to be made simultaneously.

For this to be remotely secure, rate-limiting the entire set is necessary, and
sanity checking TOTP code lengths is recommended.

This library covers both.
