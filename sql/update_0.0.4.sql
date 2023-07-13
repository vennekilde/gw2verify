
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
    ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id");

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
    ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id");

UPDATE service_links as t
SET user_id=subquery.user_id
FROM (SELECT id, user_id
      FROM accounts) AS subquery
WHERE t.account_id=subquery.id;

ALTER TABLE "service_links" DROP "account_id";


-- map all bans to a user id
ALTER TABLE "bans"
    ADD "user_id" smallint NULL,
    ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id");

UPDATE bans as t
SET user_id=subquery.user_id
FROM (SELECT id, user_id
      FROM accounts) AS subquery
WHERE t.account_id=subquery.id;


-- map token permissions to jsonb
ALTER TABLE "token_infos"
    ALTER "permissions" TYPE character varying(255);
UPDATE token_infos SET permissions = regexp_replace(permissions, '{', '["');
UPDATE token_infos SET permissions = regexp_replace(permissions, '}', '"]');
UPDATE token_infos SET permissions = regexp_replace(permissions, ', ', '","');
ALTER TABLE "token_infos"
    ALTER "permissions" TYPE jsonb USING permissions::jsonb;

-- map account guilds to jsonb
ALTER TABLE "accounts" 
    ALTER "guilds" TYPE character varying(255);
UPDATE accounts SET guilds = regexp_replace(guilds, '{', '["');
UPDATE accounts SET guilds = regexp_replace(guilds, '}', '"]');
UPDATE accounts SET guilds = regexp_replace(guilds, ', ', '","');
UPDATE accounts SET guilds = regexp_replace(guilds, ', ?', '","');
ALTER TABLE "accounts"
    ALTER "guilds" TYPE jsonb USING guilds::jsonb;
UPDATE accounts SET guilds = '[]' WHERE CAST("guilds" AS text) = '[""]' ;

-- map account guilds to jsonb
ALTER TABLE "accounts" 
    ALTER "guild_leader" TYPE character varying(255);
UPDATE accounts SET guild_leader = regexp_replace(guild_leader, '{', '["');
UPDATE accounts SET guild_leader = regexp_replace(guild_leader, '}', '"]');
UPDATE accounts SET guild_leader = regexp_replace(guild_leader, ', ?', '","');
ALTER TABLE "accounts"
    ALTER "guild_leader" TYPE jsonb USING guild_leader::jsonb;

-- map account access to jsonb
ALTER TABLE "accounts" 
    ALTER "access" TYPE character varying(255);
UPDATE accounts SET access = regexp_replace(access, '{', '["');
UPDATE accounts SET access = regexp_replace(access, '}', '"]');
UPDATE accounts SET access = regexp_replace(access, ', ', '","');
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
