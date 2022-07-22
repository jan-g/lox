# Yet another lox(ish) implementation

This is a quick (and ugly) implementation
of `lox`, [as described by Robert Nystrom](https://craftinginterpreters.com/).

There are a few bits that deviate from `lox` as written:

- comparators are nonassociative
- the `class X < X {}` limitation is removed
- function literals (both named and anonymous) are permitted

Tree-walker only at the moment.