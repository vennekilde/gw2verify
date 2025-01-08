
CREATE INDEX "achievements_account_id_achievement_timestamp" ON "public"."achievements" USING btree ("account_id", "achievement", "timestamp");
CREATE INDEX "achievements_achievement" ON "achievements" ("achievement");