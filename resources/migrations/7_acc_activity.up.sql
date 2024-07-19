TRUNCATE TABLE "activities";
ALTER TABLE "activities" ALTER "user_id" TYPE character varying(64);
ALTER TABLE "activities" RENAME "user_id" TO "account_id";