---
title: pd
section: 1
footer: pd VERSION_PLACEHOLDER
header: pd User's Manual
author: Daniel Schauenberg <d@unwiredcouch.com>
date: DATE_PLACEHOLDER
---

<!-- This is the sniple(1) man page, written in Markdown. -->
<!-- To generate the roff version, run `make man` -->

# NAME

pd — command line helper tools for PagerDuty


# SYNOPSIS

`pd incidents --teamid=<id> --since=<timeago>`

`pd oncall --user-id=STRING --since=STRING`




# EXAMPLES

`pd incidents`
: Get incidents from pagerduty API or csv.

`pd oncall`
: Retrieve on-call schedules for user from the pagerduty API.


# DESCRIPTION

`pd` is a command line tool to interact with the PagerDuty service. It
provides convenience methods to get recent incidents for a team as well as
recent on-call shifts for a user.


# META OPTIONS AND COMMANDS

`--help`
: Show list of command-line options.

`version`
: Show version of pd.



# AUTHOR

pd is maintained by mrtazz.

**Source code:** `https://github.com/mrtazz/pd`

# REPORTING BUGS

- https://github.com/mrtazz/pd/issues
