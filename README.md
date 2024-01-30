# pd
pagerduty helper tooling.

This is a small set of helpful commands to retrieve some data from PagerDuty
in markdown format to put in notes or GitHub issues.

## Usage

### General usage

The tool expects `PD_TOKEN` to contain the PagerDuty API token:
```
% export PD_TOKEN=xxxxx
```


```
% pd --help
Usage: pd <command>

Flags:
  -h, --help    Show context-sensitive help.

Commands:
  incidents
    get incidents from pagerduty API or csv

  oncall --user-id=STRING --since=STRING
    retrieve on-call schedules for user from the pagerduty API

  version
    print version and exit.

Run "pd <command> --help" for more information on a command.
```

## Get incidents for a team
```
% pd incidents --help
Usage: pd incidents

get incidents from pagerduty API or csv

Flags:
  -h, --help                  Show context-sensitive help.

      --team-id=STRING        PagerDuty team ID to get incidents for
      --since="2006-01-02"    time range to get incidents for
      --csv=STRING            PagerDuty incidents .csv export to use


% pd incidents --team-id=P789 --since=2023-10-30
|                      INCIDENT              | DESCRIPTION                         |             CREATED              | LAST UPDATE                      | DURATION |
|--------------------------------------------|-------------------------------------|----------------------------------|----------------------------------|----------|
| [123](https://pagerduty.com/incidents/123) | *api/ping is CRITICAL*              | 2023-09-20 10:00 UTC (11:00 CET) | 2023-09-20 10:18 UTC (11:18 CET) | 17m59s   |
| [456](https://pagerduty.com/incidents/456) | *database/healthcheck is CRITICAL*  | 2023-09-20 11:00 UTC (12:00 CET) | 2023-09-20 11:37 UTC (12:37 CET) | 37m1s    |
```

## Get times a user has been on-call

```
% pd oncall --help
Usage: pd oncall --user-id=STRING --since=STRING

retrieve on-call schedules for user from the pagerduty API

Flags:
  -h, --help              Show context-sensitive help.

      --user-id=STRING    PagerDuty user ID to get incidents for
      --since=STRING      time range to get incidents for, e.g. 2006-01-02

% pd oncall --user-id=P123 --since=2023-09-05
Getting on-call times for user 'mrtazz', with ID 'P123' since 2023-09-05 00:00:00 +0000 UTC:
|   START    |    END     |            SCHEDULE            |
|------------|------------|--------------------------------|
| 2023-09-11 | 2023-09-18 | on-call-L2                     |
| 2023-09-18 | 2023-09-25 | on-call-L1                     |
```
