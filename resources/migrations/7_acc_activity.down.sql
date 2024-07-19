TRUNCATE TABLE "activities";
ALTER TABLE "activities" RENAME "account_id" TO "user_id";
ALTER TABLE "activities" ALTER "user_id" TYPE integer;