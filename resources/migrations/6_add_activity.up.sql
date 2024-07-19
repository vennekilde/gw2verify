CREATE TABLE "activities" (
    "id" serial NOT NULL,
    "user_id" int NOT NULL,
    "timestamp" timestamptz DEFAULT CURRENT_TIMESTAMP NOT NULL,
    "rank" smallint NOT NULL,
    "kills" int NOT NULL,
    PRIMARY KEY ("id")
);

ALTER TABLE "accounts" ALTER "user_id" TYPE integer;
ALTER TABLE "bans" ALTER "user_id" TYPE integer;
ALTER TABLE "ephemeral_associations" ALTER "user_id" TYPE integer;
ALTER TABLE "platform_links" ALTER "user_id" TYPE integer;
ALTER TABLE "users" ALTER "id" TYPE integer;