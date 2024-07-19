CREATE TABLE "activities" (
    "id" serial NOT NULL,
    "account_id" uuid NOT NULL,
    "timestamp" timestamptz DEFAULT CURRENT_TIMESTAMP NOT NULL,
    "rank" smallint NOT NULL,
    "kills" int NOT NULL,
    PRIMARY KEY ("id")
);

DROP TABLE achievements;
DROP INDEX "achievements_account_id_achievement"