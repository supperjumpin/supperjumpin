# Supabase Staging Setup

Use Supabase as hosted staging infrastructure, not as the default local development database.

## Environment Roles

| Environment | Database | Default Use |
| --- | --- | --- |
| Local | Docker Compose Postgres | Daily API and mobile development |
| Staging | Supabase project `supperjumpin-staging` | Device testing, shared demos, hosted auth |
| Production | Separate future Supabase project | Real player data only |

## Local Development

Use local Postgres unless you are intentionally testing staging:

```sh
npm run db:reset
npm run api:dev
```

`npm run db:reset` is local-only. It starts Docker Compose Postgres, recreates the local `supperjumpin` database, and reapplies migrations.

## Staging Migrations

For Supabase from WSL or any IPv4-only network, use the **Session pooler** connection string. Supabase direct database connections may resolve to IPv6 and fail with `network is unreachable`.

```sh
DATABASE_URL='postgresql://postgres.<project-ref>:<password>@aws-0-<region>.pooler.supabase.com:5432/postgres?sslmode=require' npm run db:migrate
```

Verify migrations in Supabase SQL Editor:

```sql
select table_schema, table_name
from information_schema.tables
where table_schema = 'public'
order by table_name;

select *
from schema_migrations;
```

## API Auth

The API accepts real Supabase Auth access tokens when `SUPABASE_JWT_SECRET` is set:

```sh
DATABASE_URL='postgresql://...' SUPABASE_JWT_SECRET='...' npm run api:dev
```

`SUPABASE_JWT_SECRET` is the project's JWT signing secret from Supabase auth/API settings. It is not the client publishable key and not the new Supabase secret API key.

The local dev token still works when `SUPPERJUMPIN_DEV_AUTH_TOKEN` is set. `npm run api:dev` defaults it to `dev-token` for local smoke testing.

## Mobile Env

`apps/mobile/.env` needs:

```sh
EXPO_PUBLIC_SUPABASE_URL=https://<project-ref>.supabase.co
EXPO_PUBLIC_SUPABASE_ANON_KEY=sb_publishable_...
EXPO_PUBLIC_API_BASE_URL=http://<local-api-host>:8080
```

Supabase now calls client-safe keys **publishable keys**. The app still uses the legacy env var name `EXPO_PUBLIC_SUPABASE_ANON_KEY`; put the publishable key there. Never put a secret key in an `EXPO_PUBLIC_` variable.

## RLS

Application tables should have Row Level Security enabled with no permissive client policies unless direct Supabase table access is intentionally introduced. Tracked in GitHub issue #258.
