DROP TABLE IF EXISTS "bans";
CREATE TABLE "public"."bans" (
    "db_created" timestamptz DEFAULT now() NOT NULL,
    "db_updated" timestamptz DEFAULT now() NOT NULL,
    "account_id" character varying(64) NOT NULL,
    "expires" timestamptz default '3000-01-01 00:00:00.000000+00' NOT NULL,
    "reason" text NOT NULL
) WITH (oids = false);
