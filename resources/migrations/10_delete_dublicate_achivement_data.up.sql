
CREATE INDEX "achievements_account_id_achievement_timestamp" ON "public"."achievements" USING btree ("account_id", "achievement", "timestamp");

ALTER TABLE "achievements" ADD FOREIGN KEY ("achievement") REFERENCES "achievement_names" ("id") ON DELETE NO ACTION ON UPDATE NO ACTION;