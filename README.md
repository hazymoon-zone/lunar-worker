# lunar-worker

Worker service for the Lunar Reminder application.

It runs scheduled jobs to:
- renew repeatable reminders (monthly/yearly lunar recurrence),
- create daily alerts for eligible reminders,
- send alert emails,
- handle alert acknowledgement.

## Tech stack

- Go
- [Encore](https://encore.dev) (APIs + Cron jobs + secret management)
- PostgreSQL
- [Bob](https://github.com/stephenafamo/bob) query builder
- Mailgun (email delivery)
- lunar-go (lunar/solar date conversion)

## High-level flow

1. Reminder renewal (daily)
- Finds recurring reminders that are already past due.
- Calculates their next occurrence based on lunar recurrence rules.
- Stores the newly computed next alert date.

2. Daily alert preparation (daily)
- Cleans up old alert records.
- Finds reminders that are currently eligible for notification.
- Creates one alert record per eligible reminder for the current day.

3. Email delivery (every 2 hours)
- Loads pending alerts that have not been acknowledged.
- Sends reminder emails through Mailgun.
- Includes an acknowledgement link in each email.

4. Acknowledgement handling (public endpoint)
- Validates acknowledgement requests using a signed token.
- Marks the corresponding alert as acknowledged in the database.

## Project structure

- `/api`: cron endpoints and public acknowledge endpoint
- `/event`: domain logic (eligibility, lunar recurrence, alert creation)
- `/db`: database queries and persistence
- `/message`: Mailgun email + JWT token generation/verification
- `/lunar`: helper functions for lunar recurrence calculations

## Required secrets

Configure these Encore secrets:

- `LunarReminderDatabase` (PostgreSQL DSN)
- `MailgunApiKey`
- `MailgunSandBox` (Mailgun domain)
- `JwtSecret`

## Expected database tables

This service expects at least:

- `reminder`
  - `id`
  - `reminder_date`
  - `next_alert_date`
  - `repeat` (`monthly` or `yearly` for repeatable reminders)
  - `alert_before`
  - `mail_subject`
  - `mail_body`
- `alert`
  - `id`
  - `reminder_id`
  - `alert_date`
  - `acknowledged`

## Run locally

1. Install Encore CLI.
2. Set required secrets for local environment.
3. Run:

```bash
encore run
```

Encore will run APIs and cron jobs locally.

## Notes

- Dates written to alert/reminder scheduling are normalized to date boundaries.
- Reminder renewal logic uses lunar-go conversions to find next monthly/yearly occurrence.
- Logging is done with `encore.dev/rlog`.
