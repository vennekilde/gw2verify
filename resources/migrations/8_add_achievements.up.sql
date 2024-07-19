CREATE TABLE "achievements" (
    "id" serial NOT NULL,
    "account_id" uuid NOT NULL,
    "timestamp" timestamptz DEFAULT CURRENT_TIMESTAMP NOT NULL,
    "achievement" int NOT NULL,
    "value" int NOT NULL,
    PRIMARY KEY ("id")
);
CREATE INDEX "achievements_achievement" ON "achievements" ("achievement");
CREATE INDEX "achievements_account_id_achievement" ON "achievements" ("account_id", "achievement");

INSERT INTO "achievements" ("account_id", "timestamp", "achievement", "value")
    SELECT "account_id", "timestamp", 283, "kills" FROM "activities";
INSERT INTO "achievements" ("account_id", "timestamp", "achievement", "value")
    SELECT "account_id", "timestamp", -1, "rank" FROM "activities";

DROP TABLE activities;

ALTER TABLE "histories" DROP CONSTRAINT "histories_account_id_fkey";
ALTER TABLE "token_infos" DROP CONSTRAINT "token_infos_account_id_fkey";

ALTER TABLE "accounts" ALTER "id" TYPE uuid USING id::uuid;
ALTER TABLE "bans" ALTER "account_id" TYPE uuid USING account_id::uuid;
ALTER TABLE "histories" ALTER "account_id" TYPE uuid USING account_id::uuid;
ALTER TABLE "token_infos" ALTER "account_id" TYPE uuid USING account_id::uuid;

ALTER TABLE "token_infos" ADD FOREIGN KEY ("account_id") REFERENCES "accounts" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;
ALTER TABLE "histories" ADD FOREIGN KEY ("account_id") REFERENCES "accounts" ("id") ON DELETE NO ACTION ON UPDATE NO ACTION;