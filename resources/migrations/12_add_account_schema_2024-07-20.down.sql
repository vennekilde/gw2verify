ALTER TABLE "accounts"
    DROP "wvw_team_id",
    DROP "last_modified",
    DROP "wvw_guild_id";
    
ALTER TABLE "accounts" RENAME "wvw_rank" TO "wv_w_rank";