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

CREATE INDEX "voice_user_states_channel_id" ON "voice_user_states" ("channel_id");
CREATE INDEX "voice_user_states_timestamp" ON "voice_user_states" ("timestamp");
CREATE INDEX "voice_user_states_timestamp_service_id_channel_id" ON "voice_user_states" ("timestamp", "service_id", "channel_id");
CREATE INDEX "voice_user_states_timestamp_wvw_rank_channel_id" ON "voice_user_states" ("timestamp" DESC, "wvw_rank", "channel_id");
CREATE INDEX "voice_user_states_wvw_rank" ON "voice_user_states" ("wvw_rank");