# Database Setup and Migration Steps

This document outlines the recommended steps to create and manage two PostgreSQL database environments:

- A **production** database that remains persistent.
- A **testing/development** database that is ephemeral and resets on each run.

These steps cover environment configuration, database bootstrapping, and migration execution.

---

## Step 1: Define Docker Compose Profiles

- **Production:**
  - Use a dedicated profile (or treat the default as production).
  - Load environment variables from a file (e.g. `.env.production`).
  - Mount a persistent volume (e.g., `postgres-data`) so that your data is preserved across container restarts.

- **Testing/Development:**
  - Use a profile like `development` or `testing` (as defined in `docker-compose.yml` and `docker-compose.testing.yml`).
  - Load environment variables from a file (e.g. `.env.development`).
  - Override volume configuration to use ephemeral storage (for example, by using a `tmpfs` mount or omitting the volume mapping) so that the database starts fresh.

---

## Step 2: Configure Environment Variables

- Create environment files for each profile:
  - **Production (`.env.production`):**
    - Define variables such as `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB` with production-ready values.
    - Set any additional parameters, e.g., `DB_DRIVER`, `DB_SOURCE` (connection string), and SSL settings if needed.

  - **Testing/Development (`.env.development`):**
    - Use less critical values that are acceptable for non-production use.
    - These values might be re-used each time the environment is created and can be less secure.

- Ensure your `docker-compose.yml` references these environment files appropriately via the `env_file` directive, so that the app, postgres, and related services get the correct credentials and connection settings.

---

## Step 3: Database Bootstrapping and Initialization

- **Initialization Scripts:**
  - Use the Docker PostgreSQL image's built-in initialization mechanism by placing SQL scripts (for creating extensions, initial schema, etc.) in a folder mounted to `/docker-entrypoint-initdb.d`.
  - Note: These scripts run only when the PostgreSQL data directory is empty (i.e. on first initialization), making them ideal for a fresh, ephemeral setup in testing.

- **Persistent Production Database:**
  - With a persistent volume, the initialization scripts will only run once. Subsequent schema changes should be handled via migrations.

- **Ephemeral Testing/Development Database:**
  - Override the persistent volume settings so that every run starts with a clean state. This can be achieved by either removing the volume mapping or by using a temporary filesystem (`tmpfs`).

---

## Step 4: Running Migrations

- **Migration Tool:**
  - Use a migration tool (such as `golang-migrate`) integrated into your application to manage schema changes.

- **Entrypoint Script:**
  - Create a startup script for the app (or a separate migration service) that performs the following steps:
    1. Wait for the PostgreSQL service to be healthy (using the healthcheck defined in docker-compose).
    2. Execute the migration command (e.g., `migrate -path /app/migrations -database $DB_SOURCE up`).
    3. Log output to confirm that migrations have been applied successfully.

- **Handling Extensions and Schema Changes:**
  - Ensure that your migration scripts include statements to create required extensions, tables, and relationships as defined in your models and DTOs.
  - For production, handle schema migration carefully, including planning for potential rollback scenarios.

---

## Step 5: Access and Verification

- **Logging In:**
  - Use `psql` or another database client to log in using the credentials defined in the appropriate environment file.
  - Examples:
    - **Production:**
      ```bash
      psql "postgresql://<PROD_USER>:<PROD_PASSWORD>@<HOST>:5432/<PROD_DB>?sslmode=require"
      ```
    - **Testing/Development:**
      ```bash
      psql "postgresql://<DEV_USER>:<DEV_PASSWORD>@<HOST>:5432/<DEV_DB>?sslmode=disable"
      ```

- **Health Checks:**
  - Verify that the health checks defined in your `docker-compose.yml` are consistently passing, indicating that both your databases and application are correctly connected.

---

## Additional Considerations

- **Backup and Security (Production):**
  - Regular backups and secure handling of credentials are crucial for your production database.
  - Consider restricting access by networking configuration and proper database user privileges.

- **Monitoring and Logging:**
  - Ensure that monitoring tools (such as PostgreSQL exporters) are enabled and configured to track database health.

- **Rollback Procedures:**
  - In production, plan and document procedures for rolling back migrations or restoring from backups if necessary.

---

## Summary

By following these steps, you can set up a robust database environment where:

- The **production database** uses a persistent volume and secure credentials for continuous operation.
- The **testing/development database** is ephemeral, ensuring a fresh start for each run, with automated initialization and migration steps.

This approach helps keep your development agile while ensuring production stability and data integrity. 