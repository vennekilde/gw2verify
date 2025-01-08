ALTER TABLE "accounts"
    ADD "wvw_team_id" integer NULL,
    ADD "last_modified" timestamptz NULL,
    ADD "wvw_guild_id" uuid NULL;

ALTER TABLE "accounts" RENAME "wv_w_rank" TO "wvw_rank";