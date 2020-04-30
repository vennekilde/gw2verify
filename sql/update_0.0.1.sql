DROP TABLE IF EXISTS "accounts";
CREATE TABLE "public"."accounts" (
    "db_created" timestamptz,
    "db_updated" timestamptz,
    "id" text NOT NULL,
    "name" text,
    "world" integer,
    "age" integer,
    "guilds" character varying(255)[],
    "guild_leader" character varying(255)[],
    "access" character varying(255)[],
    "created" text,
    "commander" boolean,
    "fractal_level" integer,
    "daily_ap" integer,
    "monthly_ap" integer,
    "wvw_rank" integer,
    CONSTRAINT "accounts_pkey" PRIMARY KEY ("id")
) WITH (oids = false);


DROP TABLE IF EXISTS "service_links";
CREATE TABLE "public"."service_links" (
    "db_created" timestamptz DEFAULT now() NOT NULL,
    "db_updated" timestamptz DEFAULT now() NOT NULL,
    "account_id" text NOT NULL,
    "service_id" integer NOT NULL,
    "service_user_id" text NOT NULL,
    "is_primary" boolean NOT NULL,
    "service_user_display_name" text,
    CONSTRAINT "service_links_pkey" PRIMARY KEY ("service_id", "service_user_id")
) WITH (oids = false);


DROP TABLE IF EXISTS "temporary_accesses";
CREATE TABLE "public"."temporary_accesses" (
    "db_created" timestamptz,
    "db_updated" timestamptz,
    "service_id" integer,
    "service_user_id" text,
    "world" integer,
    CONSTRAINT "idx_ta_service_id_service_user_id" UNIQUE ("service_id", "service_user_id")
) WITH (oids = false);


DROP TABLE IF EXISTS "token_infos";
CREATE TABLE "public"."token_infos" (
    "db_created" timestamptz DEFAULT now() NOT NULL,
    "db_updated" timestamptz DEFAULT now() NOT NULL,
    "last_success" timestamptz DEFAULT now() NOT NULL,
    "id" text NOT NULL,
    "name" text,
    "api_key" text,
    "account_id" text,
    "permissions" character varying(255)[],
    CONSTRAINT "token_infos_pkey" PRIMARY KEY ("id")
) WITH (oids = false);