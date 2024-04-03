
-- cleanup dangling tokens
DELETE FROM "token_infos" WHERE "account_id" IS NULL;

CREATE TABLE "users" (
    "id" serial NOT NULL,
    "db_created" timestamptz DEFAULT now() NOT NULL,
    "db_updated" timestamptz DEFAULT now() NOT NULL,
    PRIMARY KEY ("id")
);

-- map all accounts to a user id
ALTER TABLE "accounts"
    ADD "user_id" smallint NULL,
    ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;

with rows as (
  UPDATE accounts t
  SET    user_id = nextval('users_id_seq')
  RETURNING db_created, db_updated, user_id
)
INSERT INTO users(db_created, db_updated, id)
SELECT * FROM rows;

ALTER TABLE "accounts"
    ALTER "user_id" SET NOT NULL;


-- map all service_links to a user id
ALTER TABLE "service_links"
    ADD "user_id" smallint NULL,
    ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;

UPDATE service_links as t
    SET user_id=subquery.user_id
    FROM (SELECT id, user_id FROM accounts) AS subquery
    WHERE t.account_id=subquery.id;

ALTER TABLE "service_links" DROP "account_id";


-- map all bans to a user id
ALTER TABLE "bans"
    ADD "user_id" smallint NULL,
    ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;

UPDATE bans as t
SET user_id=subquery.user_id
FROM (SELECT id, user_id
      FROM accounts) AS subquery
WHERE t.account_id=subquery.id;


-- map token permissions to jsonb
ALTER TABLE "token_infos"
    ALTER "permissions" TYPE character varying(255);
UPDATE token_infos SET permissions = regexp_replace(permissions, '{', '["', 'g');
UPDATE token_infos SET permissions = regexp_replace(permissions, '}', '"]', 'g');
UPDATE token_infos SET permissions = regexp_replace(permissions, '([[:alpha:]]|\d), ?', '","', 'g');
ALTER TABLE "token_infos"
    ALTER "permissions" TYPE jsonb USING permissions::jsonb;

-- map account guilds to jsonb
ALTER TABLE "accounts" 
    ALTER "guilds" TYPE character varying(255);
UPDATE accounts SET guilds = regexp_replace(guilds, '{', '["', 'g');
UPDATE accounts SET guilds = regexp_replace(guilds, '}', '"]', 'g');
UPDATE accounts SET guilds = regexp_replace(guilds, ', ', '","', 'g');
UPDATE accounts SET guilds = regexp_replace(guilds, '([[:alpha:]]|\d), ?', '","', 'g');
ALTER TABLE "accounts"
    ALTER "guilds" TYPE jsonb USING guilds::jsonb;
UPDATE accounts SET guilds = '[]' WHERE CAST("guilds" AS text) = '[""]' ;

-- map account guilds to jsonb
ALTER TABLE "accounts" 
    ALTER "guild_leader" TYPE character varying(255);
UPDATE accounts SET guild_leader = regexp_replace(guild_leader, '{', '["', 'g');
UPDATE accounts SET guild_leader = regexp_replace(guild_leader, '}', '"]', 'g');
UPDATE accounts SET guild_leader = regexp_replace(guild_leader, '([[:alpha:]]|\d), ?', '","', 'g');
ALTER TABLE "accounts"
    ALTER "guild_leader" TYPE jsonb USING guild_leader::jsonb;

-- map account access to jsonb
ALTER TABLE "accounts" 
    ALTER "access" TYPE character varying(255);
UPDATE accounts SET access = regexp_replace(access, '{', '["', 'g');
UPDATE accounts SET access = regexp_replace(access, '}', '"]', 'g');
UPDATE accounts SET access = regexp_replace(access, '([[:alpha:]]|\d), ?', '","', 'g');
ALTER TABLE "accounts"
    ALTER "access" TYPE jsonb USING access::jsonb;


-- make my life easier and just accept that bun wants it to be named wv_w_rank due to WvWRank being snake cased that way
ALTER TABLE "accounts" RENAME "wvw_rank" TO "wv_w_rank";

DELETE FROM "temporary_accesses";

ALTER TABLE "accounts"
    ALTER "db_created" SET DEFAULT CURRENT_TIMESTAMP,
    ALTER "db_updated" SET DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE "bans"
    ALTER "db_created" SET DEFAULT CURRENT_TIMESTAMP,
    ALTER "db_updated" SET DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE "service_links"
    ALTER "db_created" SET DEFAULT CURRENT_TIMESTAMP,
    ALTER "db_updated" SET DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE "temporary_accesses"
    ALTER "db_created" SET DEFAULT CURRENT_TIMESTAMP,
    ALTER "db_created" SET NOT NULL,
    ALTER "db_updated" SET DEFAULT CURRENT_TIMESTAMP,
    ALTER "db_updated" SET NOT NULL;
ALTER TABLE "token_infos"
    ALTER "db_created" SET DEFAULT CURRENT_TIMESTAMP,
    ALTER "db_updated" SET DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE "users"
    ALTER "db_created" SET DEFAULT CURRENT_TIMESTAMP,
    ALTER "db_updated" SET DEFAULT CURRENT_TIMESTAMP;

ALTER TABLE "token_infos"
    ALTER "name" SET NOT NULL,
    ALTER "api_key" SET NOT NULL,
    ALTER "account_id" SET NOT NULL,
    ALTER "permissions" SET NOT NULL;

ALTER TABLE "service_links" 
    RENAME "service_user_display_name" TO "display_name";
ALTER TABLE "service_links" 
    RENAME "is_primary" TO "primary";

ALTER TABLE "bans" 
    RENAME "expires" TO "until";

ALTER TABLE "temporary_accesses"
    ADD CONSTRAINT "temporary_accesses_service_id_service_user_id_world" UNIQUE ("service_id", "service_user_id", "world"),
    DROP CONSTRAINT "idx_ta_service_id_service_user_id";

CREATE TABLE "services" (
  "uuid" character varying(64) NOT NULL,
  "api_key" character varying(256) NOT NULL,
  "name" character varying(256) NOT NULL
);
ALTER TABLE "services"
    ADD CONSTRAINT "services_uuid" PRIMARY KEY ("uuid");

ALTER TABLE "service_links" RENAME TO "platform_links";
ALTER TABLE "platform_links" RENAME "service_id" TO "platform_id";
ALTER TABLE "platform_links" RENAME "service_user_id" TO "platform_user_id";

ALTER TABLE "temporary_accesses" RENAME "service_id" TO "platform_id";
ALTER TABLE "temporary_accesses" RENAME "service_user_id" TO "platform_user_id";

ALTER TABLE "voice_user_states" RENAME "service_id" TO "platform_id";
ALTER TABLE "voice_user_states" RENAME "service_user_id" TO "platform_user_id";

CREATE TABLE "properties" (
  "db_created" timestamptz NOT NULL DEFAULT NOW(),
  "db_updated" timestamptz NOT NULL DEFAULT NOW(),
  "service_uuid" character varying(64) NOT NULL,
  "subject" character varying(256) NOT NULL,
  "name" character varying(256) NOT NULL,
  "value" character varying(1024) NOT NULL
);
ALTER TABLE "properties"
    ADD CONSTRAINT "properties_service_uuid_subject_name" PRIMARY KEY ("service_uuid", "subject", "name"),
    ADD FOREIGN KEY ("service_uuid") REFERENCES "services" ("uuid") ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE "token_infos"
    ADD FOREIGN KEY ("account_id") REFERENCES "accounts" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE "histories"
    ADD FOREIGN KEY ("account_id") REFERENCES "accounts" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE "temporary_accesses"
    ADD "user_id" smallint NULL,
    ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;

UPDATE temporary_accesses as t
    SET user_id=subquery.user_id
    FROM (SELECT platform_id, platform_user_id, user_id FROM platform_links) AS subquery
    WHERE t.platform_id = subquery.platform_id AND t.platform_user_id = subquery.platform_user_id;

ALTER TABLE "temporary_accesses"
    DROP "platform_id",
    DROP "platform_user_id";

ALTER TABLE "temporary_accesses" RENAME TO "ephemeral_associations";

ALTER TABLE "ephemeral_associations"
    ADD "until" timestamptz NULL;

UPDATE "ephemeral_associations" as t
    SET until = db_updated + 1814400 * interval '1 second';

ALTER TABLE "ephemeral_associations"
    ALTER "user_id" SET NOT NULL,
    ALTER "until" SET NOT NULL;