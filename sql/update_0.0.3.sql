DROP TABLE IF EXISTS "voice_user_states";
CREATE TABLE "public"."voice_user_states" (
    "timestamp" timestamptz DEFAULT now() NOT NULL,
    "service_id" int NOT NULL,
    "service_user_id" character varying(64) NOT NULL,
    "channel_id" character varying(64) NOT NULL,
    "muted" boolean default false NOT NULL,
    "deafened" boolean default false NOT NULL,
    "streaming" boolean default false NOT NULL,
    "wvw_rank" int default 0 NOT NULL,
    "age" int default 0 NOT NULL,
    "verification_status" int NOT NULL
) WITH (oids = false);